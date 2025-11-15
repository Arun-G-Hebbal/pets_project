package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"pets_project/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims struct defines the data we store in the JWT
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// --- Password Hashing ---
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// --- JWT Generation ---
func generateJWT(userID int) (string, error) {
	jwtSecretKey := []byte(os.Getenv("JWT_SECRET"))
	expirationTime := time.Now().Add(3 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

// --- Middleware ---
func (env *Env) JwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			Warn("Unauthorized request: Missing Authorization header")
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			Warn("Invalid Authorization header format")
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := headerParts[1]
		jwtSecretKey := []byte(os.Getenv("JWT_SECRET"))
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			Warn("Invalid or expired JWT token: %v", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		Info("Authenticated request from user ID %d", claims.UserID)
		ctx := context.WithValue(r.Context(), "userID", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// --- Signup Handler ---
func (env *Env) SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if creds.Email == "" || creds.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		Warn("Signup failed: Missing email or password")
		return
	}

	hashedPassword, err := hashPassword(creds.Password)
	if err != nil {
		Error("Failed to hash password: %v", err)
		http.Error(w, "Failed to process signup", http.StatusInternalServerError)
		return
	}

	sqlStatement := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`
	var userID int
	err = env.DB.QueryRow(sqlStatement, creds.Email, hashedPassword).Scan(&userID)
	if err != nil {
		Error("Signup failed for email %s: %v", creds.Email, err)
		http.Error(w, "Email already in use or database error", http.StatusInternalServerError)
		return
	}

	Info("User %s registered successfully (ID: %d)", creds.Email, userID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User created successfully",
		"user_id": userID,
	})
}

// --- Login Handler ---
func (env *Env) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User
	sqlStatement := `SELECT id, email, password_hash FROM users WHERE email = $1`
	err := env.DB.QueryRow(sqlStatement, creds.Email).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			Warn("Login failed: User not found (%s)", creds.Email)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		} else {
			Error("Database error during login: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if !checkPasswordHash(creds.Password, user.PasswordHash) {
		Warn("Login failed: Incorrect password for %s", creds.Email)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	tokenString, err := generateJWT(user.ID)
	if err != nil {
		Error("Failed to generate JWT: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	Info("User %s logged in successfully", creds.Email)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
