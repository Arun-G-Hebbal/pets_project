package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"pets_project/internal/models"
)

// ================================
// LIST FILES FOR PET
// GET /files?pet_id=1
// ================================
func (env *Env) ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	petID := r.URL.Query().Get("pet_id")
	if petID == "" {
		http.Error(w, "pet_id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(petID)
	if err != nil {
		http.Error(w, "Invalid pet_id", http.StatusBadRequest)
		return
	}

	rows, err := env.DB.Query(`SELECT id, pet_id, file_name, file_path, uploaded_at FROM file_records WHERE pet_id = $1`, id)
	if err != nil {
		Error("Database error while fetching files: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	files := []models.FileRecord{}

	for rows.Next() {
		var fr models.FileRecord
		err := rows.Scan(&fr.ID, &fr.PetID, &fr.FileName, &fr.FilePath, &fr.UploadedAt)
		if err != nil {
			Error("Error scanning file record: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		files = append(files, fr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// ================================
// DELETE FILE
// DELETE /files/delete?id=1
// ================================
func (env *Env) DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(fileID)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	// First fetch file path
	var filePath string
	err = env.DB.QueryRow(`SELECT file_path FROM file_records WHERE id = $1`, id).Scan(&filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Delete DB record
	_, err = env.DB.Exec(`DELETE FROM file_records WHERE id = $1`, id)
	if err != nil {
		Error("Failed to delete file record: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Delete physical file
	err = os.Remove(filePath)
	if err != nil {
		Warn("Could not delete file from disk: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "File deleted successfully",
	})
}
