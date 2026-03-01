package user

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxRead = 1024 * 1024
)

type UserService struct {
	Database  UserDatabase
	SecretKey string
}

// User defines the structure of a user in the system.
type User struct {
	ID             string
	Email          string
	ValidatedEmail bool
	Username       string
	HashedPassword []byte
}

func generateToken(u *User, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  u.ID,
		"name": u.Username,
		"exp":  jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupResponse struct {
	AccessToken string `json:"accessToken"`
}

func validSignupRequest(req *SignupRequest) string {
	if req.Username == "" {
		return "username is required"
	}
	if req.Email == "" {
		return "email is required"
	}
	if req.Password == "" {
		return "password is required"
	}
	// check for the password strength:
	// - at least 8 characters
	// - at least one uppercase letter
	// - at least one lowercase letter
	// - at least one number
	// - at least one special character
	if len(req.Password) < 8 {
		return "password must be at least 8 characters long"
	}
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, c := range req.Password {
		switch {
		case 'A' <= c && c <= 'Z':
			hasUpper = true
		case 'a' <= c && c <= 'z':
			hasLower = true
		case '0' <= c && c <= '9':
			hasNumber = true
		case (c >= 33 && c <= 47) || (c >= 58 && c <= 64) || (c >= 91 && c <= 96) || (c >= 123 && c <= 126): // TODO: verify this range of special characters
			hasSpecial = true
		}
	}
	if !hasUpper {
		return "password must contain at least one uppercase letter"
	}
	if !hasLower {
		return "password must contain at least one lowercase letter"
	}
	if !hasNumber {
		return "password must contain at least one number"
	}
	if !hasSpecial {
		return "password must contain at least one special character"
	}
	return ""
}

func (h *UserService) Signup(w http.ResponseWriter, r *http.Request) {
	bodyReader := http.MaxBytesReader(w, r.Body, maxRead)
	defer func() {
		_ = bodyReader.Close()
	}()

	req := SignupRequest{}
	err := json.NewDecoder(bodyReader).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if errReason := validSignupRequest(&req); errReason != "" {
		http.Error(w, errReason, http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}
	userID := uuid.New().String()

	user := &User{
		ID:             userID,
		Email:          req.Email,
		ValidatedEmail: false, // TODO: implement email validation
		Username:       req.Username,
		HashedPassword: hashedPassword,
	}
	err = h.Database.WriteUser(user)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := generateToken(user, h.SecretKey)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(SignupResponse{AccessToken: token})
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
}

func validLoginRequest(req *LoginRequest) string {
	if req.Username == "" {
		return "username is required"
	}
	if req.Password == "" {
		return "password is required"
	}
	return ""
}

func (h *UserService) Login(w http.ResponseWriter, r *http.Request) {
	bodyReader := http.MaxBytesReader(w, r.Body, maxRead)
	defer func() {
		_ = bodyReader.Close()
	}()

	req := LoginRequest{}
	err := json.NewDecoder(bodyReader).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if errReason := validLoginRequest(&req); errReason != "" {
		http.Error(w, errReason, http.StatusBadRequest)
		return
	}

	user, err := h.Database.GetUser(req.Username)
	if err != nil {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(req.Password))
	if err != nil {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user, h.SecretKey)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(LoginResponse{AccessToken: token})
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
