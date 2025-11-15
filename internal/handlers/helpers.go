package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// getIDFromPath gets the ID from a URL path like "/pets/1"
// This function is internal to the 'handlers' package (lowercase 'g')
func getIDFromPath(w http.ResponseWriter, r *http.Request, basePath string) (int, error) {
	// basePath will be "/pets/", "/owners/", etc.
	path := r.URL.Path

	if len(path) <= len(basePath) {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return 0, fmt.Errorf("invalid ID")
	}

	idStr := strings.TrimPrefix(path, basePath)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID in path", http.StatusBadRequest)
		return 0, err
	}
	return id, nil
}
