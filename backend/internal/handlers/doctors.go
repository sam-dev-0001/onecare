package handlers

import (
	"database/sql"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/onecare/backend/internal/models"
	"github.com/onecare/backend/pkg/response"
	"golang.org/x/crypto/bcrypt"
)

type DoctorHandler struct {
	db *sql.DB
}

func NewDoctorHandler(db *sql.DB) *DoctorHandler {
	return &DoctorHandler{db: db}
}

// GET /api/v1/doctors
func (h *DoctorHandler) List(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, first_name, last_name, email, phone, gender,
		       medical_license_no, years_of_experience, bio,
		       consultation_fee, rating, is_active, created_at, updated_at
		FROM doctors WHERE is_active = 1 ORDER BY first_name`)
	if err != nil {
		response.InternalError(c, "failed to fetch doctors")
		return
	}
	defer rows.Close()

	doctors := []models.Doctor{}
	for rows.Next() {
		var d models.Doctor
		if err := rows.Scan(
			&d.ID, &d.FirstName, &d.LastName, &d.Email, &d.Phone, &d.Gender,
			&d.MedicalLicenseNo, &d.YearsOfExperience, &d.Bio,
			&d.ConsultationFee, &d.Rating, &d.IsActive, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			continue
		}
		doctors = append(doctors, d)
	}
	response.OK(c, "doctors retrieved", doctors)
}

// GET /api/v1/doctors/:id
func (h *DoctorHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid doctor id")
		return
	}

	var d models.Doctor
	err = h.db.QueryRow(`
		SELECT id, first_name, last_name, email, phone, gender,
		       medical_license_no, years_of_experience, bio,
		       consultation_fee, rating, is_active, created_at, updated_at
		FROM doctors WHERE id = ?`, id).Scan(
		&d.ID, &d.FirstName, &d.LastName, &d.Email, &d.Phone, &d.Gender,
		&d.MedicalLicenseNo, &d.YearsOfExperience, &d.Bio,
		&d.ConsultationFee, &d.Rating, &d.IsActive, &d.CreatedAt, &d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		response.NotFound(c, "doctor not found")
		return
	}
	if err != nil {
		response.InternalError(c, "database error")
		return
	}
	response.OK(c, "doctor retrieved", d)
}

// POST /api/v1/doctors  (admin only)
func (h *DoctorHandler) Create(c *gin.Context) {
	var req models.CreateDoctorRequest
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
		  (first_name, last_name, email, password_hash, phone, gender,
		   medical_license_no, years_of_experience, bio, consultation_fee)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.FirstName, req.LastName, req.Email, string(hash),
		req.Phone, req.Gender, req.MedicalLicenseNo,
		req.YearsOfExperience, req.Bio, req.ConsultationFee,
	)
	if err != nil {
		response.Conflict(c, "email or license number already exists")
		return
	}

	insertedID, _ := result.LastInsertId()

	// Link specializations
	for _, specID := range req.SpecializationIDs {
		h.db.Exec(`INSERT IGNORE INTO doctor_specialization_map (doctor_id, specialization_id) VALUES (?, ?)`,
			insertedID, specID)
	}

	response.Created(c, "doctor created", gin.H{"id": insertedID})
}

// PUT /api/v1/doctors/:id
func (h *DoctorHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid doctor id")
		return
	}

	// Only the doctor themselves or admin can update
	callerID, _ := c.Get("user_id")
	callerRole, _ := c.Get("role")
	if callerRole.(string) != "admin" && callerID.(int) != id {
		response.Unauthorized(c, "cannot update another doctor's profile")
		return
	}

	var req struct {
		FirstName         *string  `json:"first_name"`
		LastName          *string  `json:"last_name"`
		Phone             *string  `json:"phone"`
		Bio               *string  `json:"bio"`
		ConsultationFee   *float64 `json:"consultation_fee"`
		ProfilePhotoURL   *string  `json:"profile_photo_url"`
		YearsOfExperience *int     `json:"years_of_experience"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	_, err = h.db.Exec(`
		UPDATE doctors SET
		  first_name = COALESCE(?, first_name),
		  last_name  = COALESCE(?, last_name),
		  phone      = COALESCE(?, phone),
		  bio        = COALESCE(?, bio),
		  consultation_fee   = COALESCE(?, consultation_fee),
		  profile_photo_url  = COALESCE(?, profile_photo_url),
		  years_of_experience = COALESCE(?, years_of_experience)
		WHERE id = ?`,
		req.FirstName, req.LastName, req.Phone, req.Bio,
		req.ConsultationFee, req.ProfilePhotoURL, req.YearsOfExperience, id,
	)
	if err != nil {
		response.InternalError(c, "update failed")
		return
	}
	response.OK(c, "doctor updated", nil)
}

// DELETE /api/v1/doctors/:id  (soft delete, admin only)
func (h *DoctorHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid doctor id")
		return
	}
	_, err = h.db.Exec(`UPDATE doctors SET is_active = 0 WHERE id = ?`, id)
	if err != nil {
		response.InternalError(c, "delete failed")
		return
	}
	response.OK(c, "doctor deactivated", nil)
}
