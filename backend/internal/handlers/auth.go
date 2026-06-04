package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onecare/backend/internal/config"
	"github.com/onecare/backend/internal/models"
	"github.com/onecare/backend/internal/utils"
	"github.com/onecare/backend/pkg/response"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db  *sql.DB
	cfg *config.Config
}

func NewAuthHandler(db *sql.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

// ─────────────────────────────────────────────────────────────
// POST /api/v1/auth/login
// Body: { "email": "...", "password": "...", "role": "doctor|patient|admin" }
// ─────────────────────────────────────────────────────────────
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var (
		id           int
		passwordHash string
		firstName    string
		lastName     string
	)

	var query string
	switch req.Role {
	case "doctor":
		query = `SELECT id, password_hash, first_name, last_name FROM doctors WHERE email = ? AND is_active = 1`
	case "patient":
		query = `SELECT id, password_hash, first_name, last_name FROM patients WHERE email = ? AND is_active = 1`
	case "admin":
		query = `SELECT id, password_hash, first_name, last_name FROM admins WHERE email = ? AND is_active = 1`
	default:
		response.BadRequest(c, "role must be doctor, patient, or admin")
		return
	}

	err := h.db.QueryRow(query, req.Email).Scan(&id, &passwordHash, &firstName, &lastName)
	if err == sql.ErrNoRows {
		response.Unauthorized(c, "no account found with that email and role")
		return
	}
	if err != nil {
		response.InternalError(c, "database error: "+err.Error())
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		response.Unauthorized(c, "wrong password")
		return
	}

	token, expires, err := utils.GenerateToken(id, req.Role, h.cfg.JWTSecret, h.cfg.JWTExpiryHours)
	if err != nil {
		response.InternalError(c, "could not generate token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": models.LoginResponse{
			Token:   token,
			Expires: expires,
			User: gin.H{
				"id":         id,
				"first_name": firstName,
				"last_name":  lastName,
				"role":       req.Role,
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// POST /api/v1/auth/register/doctor
// Registers a new doctor account
// ─────────────────────────────────────────────────────────────
func (h *AuthHandler) RegisterDoctor(c *gin.Context) {
	var req struct {
		FirstName        string  `json:"first_name"         binding:"required"`
		LastName         string  `json:"last_name"          binding:"required"`
		Email            string  `json:"email"              binding:"required,email"`
		Password         string  `json:"password"           binding:"required,min=6"`
		Phone            string  `json:"phone"`
		MedicalLicenseNo string  `json:"medical_license_no" binding:"required"`
		ConsultationFee  float64 `json:"consultation_fee"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(c, "failed to hash password")
		return
	}

	result, err := h.db.Exec(`
		INSERT INTO doctors
		  (first_name, last_name, email, password_hash, phone, medical_license_no, consultation_fee)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		req.FirstName, req.LastName, req.Email, string(hash),
		req.Phone, req.MedicalLicenseNo, req.ConsultationFee,
	)
	if err != nil {
		response.Conflict(c, "email or license number already registered: "+err.Error())
		return
	}

	id, _ := result.LastInsertId()
	response.Created(c, "doctor registered successfully", gin.H{
		"id":    id,
		"email": req.Email,
		"role":  "doctor",
	})
}

// ─────────────────────────────────────────────────────────────
// POST /api/v1/auth/register/admin
// Registers a new admin account
// ─────────────────────────────────────────────────────────────
func (h *AuthHandler) RegisterAdmin(c *gin.Context) {
	var req struct {
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name"  binding:"required"`
		Email     string `json:"email"      binding:"required,email"`
		Password  string `json:"password"   binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(c, "failed to hash password")
		return
	}

	result, err := h.db.Exec(`
		INSERT INTO admins (first_name, last_name, email, password_hash)
		VALUES (?, ?, ?, ?)`,
		req.FirstName, req.LastName, req.Email, string(hash),
	)
	if err != nil {
		response.Conflict(c, "email already registered: "+err.Error())
		return
	}

	id, _ := result.LastInsertId()
	response.Created(c, "admin registered successfully", gin.H{
		"id":    id,
		"email": req.Email,
		"role":  "admin",
	})
}

// ─────────────────────────────────────────────────────────────
// POST /api/v1/auth/forgot-password
// Body: { "email": "...", "user_type": "doctor|patient|admin" }
// Returns: { "reset_token": "..." } - OTP sent to email (TODO)
// ─────────────────────────────────────────────────────────────
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Verify user exists
	var userID int
	var tableColumn string
	switch req.UserType {
	case "doctor":
		tableColumn = "doctors"
	case "patient":
		tableColumn = "patients"
	case "admin":
		tableColumn = "admins"
	default:
		response.BadRequest(c, "invalid user type")
		return
	}

	query := fmt.Sprintf("SELECT id FROM %s WHERE email = ? AND is_active = 1", tableColumn)
	err := h.db.QueryRow(query, req.Email).Scan(&userID)
	if err == sql.ErrNoRows {
		// Don't reveal if email exists (security best practice)
		response.OK(c, "if email exists, otp has been sent", models.ForgotPasswordResponse{
			ResetToken: "",
		})
		return
	}
	if err != nil {
		response.InternalError(c, "database error")
		return
	}

	// Generate 32-byte random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		response.InternalError(c, "failed to generate token")
		return
	}
	resetToken := hex.EncodeToString(tokenBytes)

	// Generate 6-digit OTP
	otpBytes := make([]byte, 3)
	if _, err := rand.Read(otpBytes); err != nil {
		response.InternalError(c, "failed to generate otp")
		return
	}
	otp := fmt.Sprintf("%06d", uint32(otpBytes[0])<<16|uint32(otpBytes[1])<<8|uint32(otpBytes[2]))

	// Store in database with 15-minute expiry
	expiresAt := time.Now().Add(15 * time.Minute)
	_, err = h.db.Exec(`
		INSERT INTO password_reset_tokens (user_id, user_type, token, otp, is_used, expires_at)
		VALUES (?, ?, ?, ?, 0, ?)`,
		userID, req.UserType, resetToken, otp, expiresAt,
	)
	if err != nil {
		response.InternalError(c, "failed to create reset token")
		return
	}

	// TODO: Send OTP via email service
	// For now, just return the token (in production, store OTP and send via email)
	fmt.Println("⚠️  DEBUG OTP:", otp, "for user:", userID)

	response.OK(c, "otp has been sent to your email", models.ForgotPasswordResponse{
		ResetToken: resetToken,
	})
}

// ─────────────────────────────────────────────────────────────
// POST /api/v1/auth/reset-password
// Body: { "reset_token": "...", "otp": "123456", "new_password": "..." }
// ─────────────────────────────────────────────────────────────
func (h *AuthHandler) VerifyOTPAndResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Verify token exists, OTP matches, not used, and not expired
	var userID int
	var userType string
	var isUsed bool
	var expiresAt time.Time

	err := h.db.QueryRow(`
		SELECT user_id, user_type, is_used, expires_at
		FROM password_reset_tokens
		WHERE token = ? AND otp = ?`,
		req.ResetToken, req.OTP,
	).Scan(&userID, &userType, &isUsed, &expiresAt)

	if err == sql.ErrNoRows {
		response.Unauthorized(c, "invalid reset token or otp")
		return
	}
	if err != nil {
		response.InternalError(c, "database error")
		return
	}

	if isUsed {
		response.Unauthorized(c, "this reset token has already been used")
		return
	}
	if time.Now().After(expiresAt) {
		response.Unauthorized(c, "reset token has expired")
		return
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(c, "failed to hash password")
		return
	}

	// Update password in appropriate table
	var tableColumn string
	switch userType {
	case "doctor":
		tableColumn = "doctors"
	case "patient":
		tableColumn = "patients"
	case "admin":
		tableColumn = "admins"
	default:
		response.BadRequest(c, "invalid user type")
		return
	}

	updateQuery := fmt.Sprintf("UPDATE %s SET password_hash = ?, updated_at = NOW() WHERE id = ?", tableColumn)
	_, err = h.db.Exec(updateQuery, string(hash), userID)
	if err != nil {
		response.InternalError(c, "failed to update password")
		return
	}

	// Mark token as used
	_, err = h.db.Exec(`
		UPDATE password_reset_tokens SET is_used = 1 WHERE token = ?`,
		req.ResetToken,
	)
	if err != nil {
		response.InternalError(c, "failed to mark token as used")
		return
	}

	response.OK(c, "password reset successfully", gin.H{"message": "You can now login with your new password"})
}
