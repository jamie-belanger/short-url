package main

import (
	"errors"
	"log/slog"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type DbType int
const (
	// Data is stored in memory using a map
	DatabaseMemory DbType = iota

	// Data is stored in a SQLite database file on disk
	DatabaseSqlite
)

/*
	Creates a connection to the database
*/
func (a *application) ConnectDatabase() bool {
	a.logger.Info("Database Connect initialized")
	a.mu.Lock()
	defer a.mu.Unlock()
	defer a.logger.Info("Database Connect complete")

	switch a.dbtype {
	case DatabaseMemory:
		a.links = make(map[string]string)

	case DatabaseSqlite:
		db, err := sql.Open("sqlite3", "./data/links.db")
		if err != nil {
			a.logger.Error("Database connection failed", slog.Any("error", err))
			return false
		}

		// Configure connection pool
		db.SetMaxOpenConns(1) // SQLite works best with one connection
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(0)

		// Test the connection
		if err = db.Ping(); err != nil {
			a.logger.Error("Database ping failed", slog.Any("error", err))
			return false
		}

		a.dbconn = db
		if err = createSqliteTables(db); err != nil {
			a.logger.Error("Database create failed", slog.Any("error", err))
			return false
		}
	}

	return true
}


/*
	Creates table(s) compatible with the SQLite3 driver

	# Parameters
	
	- db (*sql.DB) pointer to the database file
*/
func createSqliteTables(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS links (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT UNIQUE NOT NULL,
			link TEXT NOT NULL,
			created DATETIME DEFAULT CURRENT_TIMESTAMP
		);`

	_, err := db.Exec(query)
	return err
}


/*
	Breaks connection to the database
*/
func (a *application) DisconnectDatabase() {
	a.logger.Info("Database Disconnecting")
	a.mu.Lock()
	defer a.mu.Unlock()
	defer a.logger.Info("Database Disconnect complete")

	switch a.dbtype {
	case DatabaseMemory:
		// nothing to do

	case DatabaseSqlite:
		if err := a.dbconn.Close(); err != nil {
			a.logger.Error("Database Close Error", slog.String("message", err.Error()))
		}
	}
}



/*
	Tests the slug to see if it's already in use or not
	
	# Parameters
	
	- slug (string) = unique slug to test

	# Returns
	
	- bool = TRUE if the slug is available; FALSE if it's already in use
*/
func (a *application) TestSlugAvailable(slug string) bool {
	a.logger.Debug("TestSlugAvailable", slog.String("slug", slug))
	a.mu.RLock()
	defer a.mu.RUnlock()

	switch a.dbtype {
	case DatabaseMemory:
		// Test for collision
		if _, ok := a.links[slug]; ok {
			return false
		}

	case DatabaseSqlite:
		query := `SELECT id FROM links WHERE slug = ?`
		row := a.dbconn.QueryRow(query, slug)
		var id int
		err := row.Scan(&id)
		if err != nil && err == sql.ErrNoRows {
			return true // slug not found, so it's available, TRUE
		}
		return false
	}

	return true
}


/*
	Inserts a new link into the database
	
	# Parameters
	
	- slug (string) = unique slug to insert the record with
	- link (string) = actual hyperlink to store
*/
func (a *application) InsertLink(slug string, link string) error {
	a.logger.Info("InsertLink", slog.String("slug", slug), slog.String("link", link))
	a.mu.Lock()
	defer a.mu.Unlock()

	switch a.dbtype {
	case DatabaseMemory:
		// Test for collision
		if _, ok := a.links[slug]; ok {
			a.logger.Error("InsertLink", slog.String("message", "slug already exists"), slog.String("slug", slug))
			return errors.New("Slug already exists")
		}
		
		a.links[slug] = link
		
	case DatabaseSqlite:
		query := `INSERT INTO links (slug, link) VALUES (?, ?)`
		result, err := a.dbconn.Exec(query, slug, link)
		if err != nil {
			return err
		}
		newId, err := result.LastInsertId()
		if err != nil {
			return err
		}
		if 0 == newId {
			return errors.New("Unknown error while inserting") // possible?
		}
	}

	a.logger.Info("InsertLink", slog.String("message", "Link inserted"))
	return nil
}

/*
	Deletes a link from the database
	
	# Parameters
	
	- slug (string) = unique slug to delete
*/
func (a *application) RemoveLink(slug string) error {
	a.logger.Info("RemoveLink", slog.String("slug", slug))
	a.mu.Lock()
	defer a.mu.Unlock()

	switch a.dbtype {
	case DatabaseMemory:
		// Test for collision
		if _, ok := a.links[slug]; !ok {
			a.logger.Error("RemoveLink", slog.String("message", "slug does not exist"), slog.String("slug", slug))
			return errors.New("Slug does not exist")
		}
		delete(a.links, slug)
		
	case DatabaseSqlite:
		query := `DELETE FROM links WHERE slug = ?`
		result, err := a.dbconn.Exec(query, slug)
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if 0 == rowsAffected {
			return errors.New("Slug not found")
		}
	}

	a.logger.Info("RemoveLink", slog.String("message", "Link deleted"))
	return nil
}

/*
	Retrieves a link from the database using the slug provided
	
	# Parameters
	
	- slug (string) = unique slug to insert the record with

	# Returns
	
	- string = the hyperlink associated with this slug

	- error = if something went wrong, details
*/
func (a *application) GetLink(slug string) (string, error) {
	a.logger.Info("GetLink", slog.String("slug", slug))
	a.mu.RLock()
	defer a.mu.RUnlock()

	switch a.dbtype {
	case DatabaseMemory:
		if link, ok := a.links[slug]; ok {
			a.logger.Info("GetLink", slog.String("message", "Link found"), slog.String("link", link))
			return link, nil
		}

	case DatabaseSqlite:
		query := `SELECT link FROM links WHERE slug = ?`
		row := a.dbconn.QueryRow(query, slug)
		var link string
		err := row.Scan(&link)
		a.logger.Info("GetLink called", slog.String("slug", slug), slog.String("link", link), slog.Any("err", err))
		if err == nil {
			a.logger.Info("GetLink", slog.String("message", "Link found"), slog.String("link", link))
			return link, nil
		}
	}

	a.logger.Error("GetLink", slog.String("message", "slug not found"), slog.String("slug", slug))
	return "", errors.New("Slug not found")
}
