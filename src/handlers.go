package main

import (
	"encoding/json"
	"net/http"
	"log/slog"
)

type errorResponse struct {
	Message string `json:"message"`
}
type createResponse struct {
	Slug string `json:"slug"`
}
type retrieveResponse struct {
	Link string `json:"link"`
}

/*
	Writes a JSON error message to the output stream, along with HTTP status code
	
	# Parameters
	
	- w (http.ResponseWriter) = reference to the current response writer

	- statusCode (int) = status code to write (ie 400, 500)

	- message (string) = what to write to the output (ie "Bad request", "Malformed input", etc)

*/
func (a *application) sendJsonErrorMessage(w http.ResponseWriter, statusCode int, message string) {
	a.logger.Info("sendJsonErrorMessage", slog.Int("statusCode", statusCode), slog.String("message", message))

	// Input validation
	if "" == message {
		message = "Unknown error"
	}
	// Note that I'm not testing every single possible value here
	switch {
	case statusCode >= 100 && statusCode < 400:
		a.logger.Warn("sendJsonErrorMessage", slog.String("reason", "called with non-error code"), slog.Int("statusCode", statusCode))
	case statusCode >= 400 && statusCode <= 599:
		// ok! well, usually
	default:
		a.logger.Error("sendJsonErrorMessage", slog.String("reason", "called with unknown or invalid code"), slog.Int("statusCode", statusCode))
		statusCode = http.StatusInternalServerError
	}

	response := errorResponse{
		Message: message,
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

/*
	Creates a shorted link
	
	# Parameters (form)

	- link (string) = the link to shorten and store
	
	# Returns
	
	string = the short slug you can use to retrieve the link later
*/
func (a *application) shortLinkCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate link in payload
	userLink := r.FormValue("link")
	if 0 == len(userLink) {
		a.sendJsonErrorMessage(w, http.StatusBadRequest, "Payload form does not contain definition for 'link' value")
		return
	}

	// Since this is intended to store hyperlinks, let's ensure we have one:
	if "http://" != userLink[:7] && "https://" != userLink[:8] {
		a.sendJsonErrorMessage(w, http.StatusBadRequest, "Provided link does not look like a valid URL")
		return
	}

	// Generate short slug. Passing a GUID in here instead of the link would make this a lot more random.
	slug := a.generateSlug(userLink)

	// Add to memory map
	a.links[slug] = userLink

	// return slug to caller
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createResponse{ Slug: slug })
}

/*
	Retrieves a stored hyperlink using the slug provided in the URL

	# Parameters (URL)

	- id (string) = a slug to retrieve

	# Returns
	
	string = the URL that is stored for this slug
*/
func (a *application) shortLinkGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// First ensure the link ID exists in path; otherwise it's a 400 Bad Request
	linkId := r.PathValue("id")
	if "" == linkId {
		a.sendJsonErrorMessage(w, http.StatusBadRequest, "ID not provided in URL")
		return
	}

	// Check the memory map for this ID
	if link, ok := a.links[linkId]; ok {
		json.NewEncoder(w).Encode(retrieveResponse{ Link: link })
	} else {
		a.sendJsonErrorMessage(w, http.StatusNotFound, "ID not found")
	}
}
