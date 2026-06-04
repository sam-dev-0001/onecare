package handlers

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/onecare/backend/internal/models"
	"github.com/onecare/backend/pkg/response"
	"golang.org/x/crypto/bcrypt"
)

type PatientHandler struct {
	db *sql.DB
}

func NewPatientHandler(db *sql.DB) *PatientHandler {
	return &PatientHandler{db: db}
}

// POST /api/v1/patients/register
func (h *PatientHandler) Register(c *gin.Context) {
	var req models.CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(c, "failed to hash password")
		return
	}

	country := "India"
	if req.Country != "" {
		country = req.Country
	}

	result, err := h.db.Exec(`
		INSERT INTO patients
		  (name, email, password_hash, phone, gender, blood_group, country)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		req.Name, req.Email, string(hash),
		req.Phone, req.Gender, req.BloodGroup, country,
	)
	if err != nil {
		response.Conflict(c, "email or phone already registered")
		return
	}

	insertedID, _ := result.LastInsertId()
	response.Created(c, "patient registered", gin.H{"id": insertedID})
}

// GET /api/v1/patients/:id
func (h *PatientHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid patient id")
		return
	}

	// Patients can only see their own profile
	callerID, _ := c.Get("user_id")
	callerRole, _ := c.Get("role")
	if callerRole.(string) == "patient" && callerID.(int) != id {
		response.Unauthorized(c, "access denied")
		return
	}

	var p models.Patient
	var allergiesJSON, conditionsJSON []byte

	err = h.db.QueryRow(`
		SELECT id, name, email, phone, gender,
		       date_of_birth, blood_group, national_id,
		       emergency_contact_name, emergency_contact_phone, emergency_contact_rel,
		       address_line1, address_line2, city, state, postal_code, country,
		       known_allergies, chronic_conditions, is_active, created_at, updated_at
		FROM patients WHERE id = ?`, id).Scan(
		&p.ID, &p.Name, &p.Email, &p.Phone, &p.Gender,
		&p.DateOfBirth, &p.BloodGroup, &p.NationalID,
		&p.EmergencyContactName, &p.EmergencyContactPhone, &p.EmergencyContactRel,
		&p.AddressLine1, &p.AddressLine2, &p.City, &p.State, &p.PostalCode, &p.Country,
		&allergiesJSON, &conditionsJSON, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		response.NotFound(c, "patient not found")
		return
	}
	if err != nil {
		response.InternalError(c, "database error")
		return
	}

	if allergiesJSON != nil {
		json.Unmarshal(allergiesJSON, &p.KnownAllergies)
	}
	if conditionsJSON != nil {
		json.Unmarshal(conditionsJSON, &p.ChronicConditions)
	}

	response.OK(c, "patient retrieved", p)
}

// GET /api/v1/patients  (doctors/admins only)
func (h *PatientHandler) List(c *gin.Context) {
	search := c.Query("search")
	query := `SELECT id, first_name, last_name, email, phone, blood_group, is_active, created_at, updated_at
		FROM patients WHERE is_active = 1`
	args := []interface{}{}

	if search != "" {
		query += ` AND MATCH(first_name, last_name) AGAINST (? IN BOOLEAN MODE)`
		args = append(args, search+"*")
	}
	query += " ORDER BY first_name LIMIT 50"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		response.InternalError(c, "failed to fetch patients")
		return
	}
	defer rows.Close()

	patients := []models.Patient{}
	for rows.Next() {
		var p models.Patient
		if err := rows.Scan(&p.ID, &p.Name, &p.Email,
			&p.Phone, &p.BloodGroup, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		patients = append(patients, p)
	}
	response.OK(c, "patients retrieved", patients)
}

// PUT /api/v1/patients/:id
// Update patient profile (name, address, etc.)
func (h *PatientHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid patient id")
		return
	}

	// Patients can only update their own profile
	callerID, _ := c.Get("user_id")
	callerRole, _ := c.Get("role")
	if callerRole.(string) == "patient" && callerID.(int) != id {
		response.Unauthorized(c, "access denied")
		return
	}

	var req models.UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Build dynamic UPDATE query based on provided fields
	updates := []string{}
	args := []interface{}{}

	if req.FirstName != nil {
		updates = append(updates, "first_name = ?")
		args = append(args, *req.FirstName)
	}
	if req.LastName != nil {
		updates = append(updates, "last_name = ?")
		args = append(args, *req.LastName)
	}
	if req.Gender != nil {
		updates = append(updates, "gender = ?")
		args = append(args, *req.Gender)
	}
	if req.DateOfBirth != nil {
		updates = append(updates, "date_of_birth = ?")
		args = append(args, *req.DateOfBirth)
	}
	if req.BloodGroup != nil {
		updates = append(updates, "blood_group = ?")
		args = append(args, *req.BloodGroup)
	}
	if req.NationalID != nil {
		updates = append(updates, "national_id = ?")
		args = append(args, *req.NationalID)
	}
	if req.ProfilePhotoURL != nil {
		updates = append(updates, "profile_photo_url = ?")
		args = append(args, *req.ProfilePhotoURL)
	}
	if req.AddressLine1 != nil {
		updates = append(updates, "address_line1 = ?")
		args = append(args, *req.AddressLine1)
	}
	if req.AddressLine2 != nil {
		updates = append(updates, "address_line2 = ?")
		args = append(args, *req.AddressLine2)
	}
	if req.City != nil {
		updates = append(updates, "city = ?")
		args = append(args, *req.City)
	}
	if req.State != nil {
		updates = append(updates, "state = ?")
		args = append(args, *req.State)
	}
	if req.PostalCode != nil {
		updates = append(updates, "postal_code = ?")
		args = append(args, *req.PostalCode)
	}
	if req.Country != nil {
		updates = append(updates, "country = ?")
		args = append(args, *req.Country)
	}

	if len(updates) == 0 {
		response.BadRequest(c, "no fields to update")
		return
	}

	// Always update the updated_at timestamp
	updates = append(updates, "updated_at = NOW()")
	args = append(args, id)

	query := "UPDATE patients SET " + strings.Join(updates, ", ") + " WHERE id = ?"
	_, err = h.db.Exec(query, args...)
	if err != nil {
		response.InternalError(c, "failed to update patient")
		return
	}

	response.OK(c, "patient profile updated", gin.H{"id": id})
}

// PUT /api/v1/patients/:id/health-info
// Update known allergies and chronic conditions
func (h *PatientHandler) UpdateHealthInfo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid patient id")
		return
	}

	// Patients can only update their own info
	callerID, _ := c.Get("user_id")
	callerRole, _ := c.Get("role")
	if callerRole.(string) == "patient" && callerID.(int) != id {
		response.Unauthorized(c, "access denied")
		return
	}

	var req models.UpdateHealthInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert arrays to JSON
	allergiesJSON, _ := json.Marshal(req.KnownAllergies)
	conditionsJSON, _ := json.Marshal(req.ChronicConditions)

	_, err = h.db.Exec(`
		UPDATE patients
		SET known_allergies = ?, chronic_conditions = ?, updated_at = NOW()
		WHERE id = ?`,
		allergiesJSON, conditionsJSON, id,
	)
	if err != nil {
		response.InternalError(c, "failed to update health info")
		return
	}

	response.OK(c, "health information updated", gin.H{"id": id})
}

// PUT /api/v1/patients/:id/emergency-contact
// Update emergency contact information
func (h *PatientHandler) UpdateEmergencyContact(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid patient id")
		return
	}

	// Patients can only update their own info
	callerID, _ := c.Get("user_id")
	callerRole, _ := c.Get("role")
	if callerRole.(string) == "patient" && callerID.(int) != id {
		response.Unauthorized(c, "access denied")
		return
	}

	var req models.UpdateEmergencyContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	_, err = h.db.Exec(`
		UPDATE patients
		SET emergency_contact_name = ?, emergency_contact_phone = ?, emergency_contact_rel = ?, updated_at = NOW()
		WHERE id = ?`,
		req.EmergencyContactName, req.EmergencyContactPhone, req.EmergencyContactRel, id,
	)
	if err != nil {
		response.InternalError(c, "failed to update emergency contact")
		return
	}

	response.OK(c, "emergency contact updated", gin.H{"id": id})
}
