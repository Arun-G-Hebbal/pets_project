package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"pets_project/internal/models" // Import your models
)

// Env struct will hold dependencies like the database connection
// This struct is shared by auth.go and handlers.go (since they are in the same package)
type Env struct {
	DB *sql.DB
}

// === Pet Handlers =================================================================
// PetsHandler (capitalized) is the mini-router. Exported to main.go.
func (env *Env) PetsHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/pets" {
		switch r.Method {
		case "GET":
			env.getAllPets(w, r)
		case "POST":
			env.createPet(w, r)
		default:
			http.Error(w, "Method not allowed for /pets", http.StatusMethodNotAllowed)
		}
	} else if strings.HasPrefix(path, "/pets/") {
		id, err := getIDFromPath(w, r, "/pets/")
		if err != nil {
			return
		}
		switch r.Method {
		case "GET":
			env.getPetByID(w, r, id)
		case "PUT":
			env.updatePet(w, r, id)
		case "DELETE":
			env.deletePet(w, r, id)
		default:
			http.Error(w, "Method not allowed for /pets/{id}", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// --- Pet CRUD Functions (internal) ---
func (env *Env) getAllPets(w http.ResponseWriter, r *http.Request) {
	rows, err := env.DB.Query("SELECT id, name, species, breed, owner_id, medical_history FROM pets")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	pets := []models.Pet{} // Use models.Pet
	for rows.Next() {
		var p models.Pet
		if err := rows.Scan(&p.ID, &p.Name, &p.Species, &p.Breed, &p.OwnerID, &p.MedicalHistory); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pets = append(pets, p)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pets)
}

func (env *Env) createPet(w http.ResponseWriter, r *http.Request) {
	var p models.Pet // Use models.Pet
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if p.Name == "" || p.OwnerID == 0 {
		http.Error(w, "Name and owner_id are required fields", http.StatusBadRequest)
		return
	}
	sqlStatement := `
		INSERT INTO pets (name, species, breed, owner_id, medical_history)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	err := env.DB.QueryRow(sqlStatement, p.Name, p.Species, p.Breed, p.OwnerID, p.MedicalHistory).Scan(&p.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (env *Env) getPetByID(w http.ResponseWriter, r *http.Request, id int) {
	var p models.Pet // Use models.Pet
	sqlStatement := `SELECT id, name, species, breed, owner_id, medical_history FROM pets WHERE id = $1`
	err := env.DB.QueryRow(sqlStatement, id).Scan(&p.ID, &p.Name, &p.Species, &p.Breed, &p.OwnerID, &p.MedicalHistory)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Pet not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (env *Env) updatePet(w http.ResponseWriter, r *http.Request, id int) {
	var p models.Pet // Use models.Pet
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sqlStatement := `
		UPDATE pets
		SET name = $1, species = $2, breed = $3, owner_id = $4, medical_history = $5
		WHERE id = $6
		RETURNING id`
	var updatedID int
	err := env.DB.QueryRow(sqlStatement, p.Name, p.Species, p.Breed, p.OwnerID, p.MedicalHistory, id).Scan(&updatedID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Pet not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	p.ID = updatedID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (env *Env) deletePet(w http.ResponseWriter, r *http.Request, id int) {
	sqlStatement := `DELETE FROM pets WHERE id = $1`
	res, err := env.DB.Exec(sqlStatement, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	count, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "Pet not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Pet deleted successfully"})
}

// === Owner Handlers =================================================================
// OwnersHandler (capitalized) is the mini-router. Exported to main.go.
func (env *Env) OwnersHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/owners" {
		switch r.Method {
		case "GET":
			env.getAllOwners(w, r)
		case "POST":
			env.createOwner(w, r)
		default:
			http.Error(w, "Method not allowed for /owners", http.StatusMethodNotAllowed)
		}
	} else if strings.HasPrefix(path, "/owners/") {
		id, err := getIDFromPath(w, r, "/owners/")
		if err != nil {
			return
		}
		switch r.Method {
		case "GET":
			env.getOwnerByID(w, r, id)
		case "PUT":
			env.updateOwner(w, r, id)
		case "DELETE":
			env.deleteOwner(w, r, id)
		default:
			http.Error(w, "Method not allowed for /owners/{id}", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// --- Owner CRUD Functions (internal) ---
func (env *Env) getAllOwners(w http.ResponseWriter, r *http.Request) {
	rows, err := env.DB.Query("SELECT id, name, contact, email FROM owners")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	owners := []models.Owner{} // Use models.Owner
	for rows.Next() {
		var o models.Owner
		if err := rows.Scan(&o.ID, &o.Name, &o.Contact, &o.Email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		owners = append(owners, o)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(owners)
}

func (env *Env) createOwner(w http.ResponseWriter, r *http.Request) {
	var o models.Owner // Use models.Owner
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if o.Name == "" || o.Email == "" {
		http.Error(w, "Name and email are required fields", http.StatusBadRequest)
		return
	}
	sqlStatement := `
		INSERT INTO owners (name, contact, email)
		VALUES ($1, $2, $3)
		RETURNING id`
	err := env.DB.QueryRow(sqlStatement, o.Name, o.Contact, o.Email).Scan(&o.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(o)
}

func (env *Env) getOwnerByID(w http.ResponseWriter, r *http.Request, id int) {
	var o models.Owner // Use models.Owner
	sqlStatement := `SELECT id, name, contact, email FROM owners WHERE id = $1`
	err := env.DB.QueryRow(sqlStatement, id).Scan(&o.ID, &o.Name, &o.Contact, &o.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Owner not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}

func (env *Env) updateOwner(w http.ResponseWriter, r *http.Request, id int) {
	var o models.Owner // Use models.Owner
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sqlStatement := `
		UPDATE owners
		SET name = $1, contact = $2, email = $3
		WHERE id = $4
		RETURNING id`
	var updatedID int
	err := env.DB.QueryRow(sqlStatement, o.Name, o.Contact, o.Email, id).Scan(&updatedID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Owner not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	o.ID = updatedID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}

func (env *Env) deleteOwner(w http.ResponseWriter, r *http.Request, id int) {
	sqlStatement := `DELETE FROM owners WHERE id = $1`
	res, err := env.DB.Exec(sqlStatement, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	count, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "Owner not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Owner deleted successfully"})
}

// === Appointment Handlers =========================================================
// AppointmentsHandler (capitalized) is the mini-router. Exported to main.go.
func (env *Env) AppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/appointments" {
		switch r.Method {
		case "GET":
			env.getAllAppointments(w, r)
		case "POST":
			env.createAppointment(w, r)
		default:
			http.Error(w, "Method not allowed for /appointments", http.StatusMethodNotAllowed)
		}
	} else if strings.HasPrefix(path, "/appointments/") {
		id, err := getIDFromPath(w, r, "/appointments/")
		if err != nil {
			return
		}
		switch r.Method {
		case "GET":
			env.getAppointmentByID(w, r, id)
		case "PUT":
			env.updateAppointment(w, r, id)
		case "DELETE":
			env.deleteAppointment(w, r, id)
		default:
			http.Error(w, "Method not allowed for /appointments/{id}", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// --- Appointment CRUD Functions (internal) ---
func (env *Env) getAllAppointments(w http.ResponseWriter, r *http.Request) {
	rows, err := env.DB.Query("SELECT id, pet_id, appointment_date, appointment_time, reason FROM appointments")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	appointments := []models.Appointment{} // Use models.Appointment
	for rows.Next() {
		var a models.Appointment
		if err := rows.Scan(&a.ID, &a.PetID, &a.AppointmentDate, &a.AppointmentTime, &a.Reason); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		appointments = append(appointments, a)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appointments)
}

func (env *Env) createAppointment(w http.ResponseWriter, r *http.Request) {
	var a models.Appointment // Use models.Appointment
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if a.PetID == 0 || a.AppointmentDate == "" || a.AppointmentTime == "" {
		http.Error(w, "pet_id, appointment_date, and appointment_time are required", http.StatusBadRequest)
		return
	}
	sqlStatement := `
		INSERT INTO appointments (pet_id, appointment_date, appointment_time, reason)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	err := env.DB.QueryRow(sqlStatement, a.PetID, a.AppointmentDate, a.AppointmentTime, a.Reason).Scan(&a.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

func (env *Env) getAppointmentByID(w http.ResponseWriter, r *http.Request, id int) {
	var a models.Appointment // Use models.Appointment
	sqlStatement := `SELECT id, pet_id, appointment_date, appointment_time, reason FROM appointments WHERE id = $1`
	err := env.DB.QueryRow(sqlStatement, id).Scan(&a.ID, &a.PetID, &a.AppointmentDate, &a.AppointmentTime, &a.Reason)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Appointment not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}

func (env *Env) updateAppointment(w http.ResponseWriter, r *http.Request, id int) {
	var a models.Appointment // Use models.Appointment
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sqlStatement := `
		UPDATE appointments
		SET pet_id = $1, appointment_date = $2, appointment_time = $3, reason = $4
		WHERE id = $5
		RETURNING id`
	var updatedID int
	err := env.DB.QueryRow(sqlStatement, a.PetID, a.AppointmentDate, a.AppointmentTime, a.Reason, id).Scan(&updatedID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Appointment not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	a.ID = updatedID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}

func (env *Env) deleteAppointment(w http.ResponseWriter, r *http.Request, id int) {
	sqlStatement := `DELETE FROM appointments WHERE id = $1`
	res, err := env.DB.Exec(sqlStatement, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	count, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "Appointment not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Appointment deleted successfully"})
}
