package models

// Pet struct corresponds to the 'pets' table
type Pet struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Species        string `json:"species"`
	Breed          string `json:"breed"`
	OwnerID        int    `json:"owner_id"`
	MedicalHistory string `json:"medical_history"`
}

// Owner struct corresponds to the 'owners' table
type Owner struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Email   string `json:"email"`
}

// Appointment struct corresponds to the 'appointments' table
type Appointment struct {
	ID              int    `json:"id"`
	PetID           int    `json:"pet_id"`
	AppointmentDate string `json:"appointment_date"`
	AppointmentTime string `json:"appointment_time"`
	Reason          string `json:"reason"`
}

// User struct corresponds to the 'users' table
type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"` // never included in JSON output
}

// Credentials struct for handling login/signup JSON data
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// FileRecord struct corresponds to 'file_records' table
type FileRecord struct {
	ID         int    `json:"id"`
	PetID      int    `json:"pet_id"`
	FileName   string `json:"file_name"`
	FilePath   string `json:"file_path"`
	UploadedAt string `json:"uploaded_at"`
}
