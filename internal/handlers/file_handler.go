package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"pets_project/internal/models"
)

// UploadFileHandler handles uploading a pet's medical record (PDF/image)
func (env *Env) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (limit 10 MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		Error("Failed to parse multipart form: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		Error("Error retrieving the file: %v", err)
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	petIDStr := r.FormValue("pet_id")
	petID, err := strconv.Atoi(petIDStr)
	if err != nil || petID <= 0 {
		Warn("Invalid pet_id provided for upload: %s", petIDStr)
		http.Error(w, "Invalid pet_id", http.StatusBadRequest)
		return
	}

	// Ensure upload directory exists
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			Error("Failed to create upload directory: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
	}

	// Create unique file name with timestamp
	fileExt := filepath.Ext(handler.Filename)
	newFileName := fmt.Sprintf("pet%d_%d%s", petID, time.Now().Unix(), fileExt)
	filePath := filepath.Join(uploadDir, newFileName)

	// Save file to server
	dst, err := os.Create(filePath)
	if err != nil {
		Error("Failed to create file on disk: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		Error("Error copying file to disk: %v", err)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// Insert metadata into DB, return id and uploaded_at
	var recordID int
	var uploadedAt time.Time
	sqlStatement := `
		INSERT INTO file_records (pet_id, file_name, file_path)
		VALUES ($1, $2, $3)
		RETURNING id, uploaded_at
	`
	err = env.DB.QueryRow(sqlStatement, petID, handler.Filename, filePath).Scan(&recordID, &uploadedAt)
	if err != nil {
		Error("DB insert failed: %v", err)
		// attempt to remove saved file if DB insert fails
		_ = os.Remove(filePath)
		http.Error(w, "Database error while saving metadata", http.StatusInternalServerError)
		return
	}

	record := models.FileRecord{
		ID:         recordID,
		PetID:      petID,
		FileName:   handler.Filename,
		FilePath:   filePath,
		UploadedAt: uploadedAt.Format(time.RFC3339),
	}

	Info("File uploaded successfully: %s (Pet ID: %d)", handler.Filename, petID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

// DownloadFileHandler allows users to download a pet’s file by ID
func (env *Env) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		Warn("Invalid file ID requested: %s", idStr)
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	var fileRecord models.FileRecord
	var uploadedAt time.Time
	sqlStatement := `SELECT id, pet_id, file_name, file_path, uploaded_at FROM file_records WHERE id = $1`
	err = env.DB.QueryRow(sqlStatement, id).Scan(&fileRecord.ID, &fileRecord.PetID, &fileRecord.FileName, &fileRecord.FilePath, &uploadedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			Warn("File not found in DB: id=%d", id)
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			Error("Error fetching file from DB: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}
	fileRecord.UploadedAt = uploadedAt.Format(time.RFC3339)

	// Verify file exists on disk
	if _, err := os.Stat(fileRecord.FilePath); os.IsNotExist(err) {
		Error("File not found on disk: %s", fileRecord.FilePath)
		http.Error(w, "File not found on server", http.StatusInternalServerError)
		return
	}

	// Serve file as attachment
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileRecord.FileName))
	http.ServeFile(w, r, fileRecord.FilePath)
	Info("File downloaded: %s (Pet ID: %d)", fileRecord.FileName, fileRecord.PetID)
}
