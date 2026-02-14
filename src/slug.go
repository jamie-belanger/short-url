package main

import (
	"crypto/sha256"
	"encoding/base64"
	"log/slog"
)

// Easily control the minimum slug length here.
const minimumSlugLength = 6

/*
	Generates a unique slug to use in lieu of the user's hyperlink
	
	# Parameters:

	- link (string): the URL to use (ie "https://www.google.com")
	
	# Returns:

	- string: the generated slug value (ie "rGu2a")
*/
func (a *application) generateSlug(link string) string {
	a.logger.Debug("generateSlug", slog.String("link", link), slog.Int("size", len(link)))
	hash := sha256.Sum256([]byte(link))
	a.logger.Debug("generateSlug", slog.Any("hash", hash), slog.Int("size", len(hash)))
	slug := base64.URLEncoding.EncodeToString(hash[:])
	a.logger.Debug("generateSlug", slog.String("slug", slug), slog.Int("size", len(slug)))

	// This is still too big (44 chars), so let's trim and return the first unique substring we see
	for i := minimumSlugLength; i < len(slug); i++ {
		if ok := a.TestSlugAvailable(slug[:i]); ok {
			return slug[:i]
		}
	}

	return slug
}
