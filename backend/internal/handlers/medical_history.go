package handlers

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onecare/backend/internal/models"
	"github.com/onecare/backend/pkg/response"
)

type MedicalHistoryHandler struct {
	db *sql.DB
}

func NewMedicalHistoryHandler(db *sql.DB) *MedicalHistoryHandler {
	return &MedicalHistoryHandler{db: db}
}

// GET /api/v1/patients/:id/records
func (h *MedicalHistoryHandler) ListForPatient(c *gin.Context) {
	patientID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid patient id")
		return
	}

	// Patients can only see non-confidential records unless it's their own
	callerID, _ := c.Get("user_id")
	callerRole, _ := c.Get("role")

	var query string
	var args []interface{}

	if callerRole.(string) == "patient" {
		if callerID.(int) != patientID {
			response.Unauthorized(c, "access denied")
			return
		}
		query = `SELECT id, patient_id, doctor_id, appointment_id, record_type, record_date,
			         chief_complaint, diagnosis, icd10_codes, symptoms, treatment_plan,
			         follow_up_required, follow_up_date, blood_pressure_sys, blood_pressure_dia,
			         heart_rate, temperature, weight_kg, height_cm, oxygen_saturation,
			         notes, is_confidential, created_at, updated_at
			  FROM medical_history WHERE patient_id = ? AND is_confidential = 0 ORDER BY record_date DESC`
		args = []interface{}{patientID}
	} else {
		// Doctors and admins see everything
		query = `SELECT id, patient_id, doctor_id, appointment_id, record_type, record_date,
			         chief_complaint, diagnosis, icd10_codes, symptoms, treatment_plan,
			         follow_up_required, follow_up_date, blood_pressure_sys, blood_pressure_dia,
			         heart_rate, temperature, weight_kg, height_cm, oxygen_saturation,
			         notes, is_confidential, created_at, updated_at
			  FROM medical_history WHERE patient_id = ? ORDER BY record_date DESC`
		args = []interface{}{patientID}
	}

	rows, err := h.db.Query(query, args...)
	if err != nil {
		response.InternalError(c, "failed to fetch records")
		return
	}
	defer rows.Close()

	records := []models.MedicalHistory{}
	for rows.Next() {
		var m models.MedicalHistory
		var icd10JSON, symptomsJSON []byte
		var followUpDate *string

		if err := rows.Scan(
			&m.ID, &m.PatientID, &m.DoctorID, &m.AppointmentID, &m.RecordType, &m.RecordDate,
			&m.ChiefComplaint, &m.Diagnosis, &icd10JSON, &symptomsJSON, &m.TreatmentPlan,
			&m.FollowUpRequired, &followUpDate, &m.BloodPressureSys, &m.BloodPressureDia,
			&m.HeartRate, &m.Temperature, &m.WeightKg, &m.HeightCm, &m.OxygenSaturation,
			&m.Notes, &m.IsConfidential, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			continue
		}

		if icd10JSON != nil {
			json.Unmarshal(icd10JSON, &m.ICD10Codes)
		}
		if symptomsJSON != nil {
			json.Unmarshal(symptomsJSON, &m.Symptoms)
		}
		if followUpDate != nil {
			t, _ := time.Parse("2006-01-02", *followUpDate)
			m.FollowUpDate = &t
		}

		records = append(records, m)
	}
	response.OK(c, "medical records retrieved", records)
}

// POST /api/v1/patients/:id/records  (doctors only)
func (h *MedicalHistoryHandler) Create(c *gin.Context) {
	patientID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid patient id")
		return
	}

	var req models.CreateMedicalHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	doctorID, _ := c.Get("user_id")

	icd10JSON, _ := json.Marshal(req.ICD10Codes)
	symptomsJSON, _ := json.Marshal(req.Symptoms)

	recordDate := time.Now().Format("2006-01-02")
	if req.RecordDate != "" {
		recordDate = req.RecordDate
	}

	var followUpDate interface{}
	if req.FollowUpDate != nil {
		followUpDate = *req.FollowUpDate
	}

	result, err := h.db.Exec(`
		INSERT INTO medical_history
		  (patient_id, doctor_id, appointment_id, record_type, record_date,
		   chief_complaint, diagnosis, icd10_codes, symptoms, treatment_plan,
		   follow_up_required, follow_up_date, blood_pressure_sys, blood_pressure_dia,
		   heart_rate, temperature, weight_kg, height_cm, oxygen_saturation,
		   notes, is_confidential)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		patientID, doctorID, req.AppointmentID, req.RecordType, recordDate,
		req.ChiefComplaint, req.Diagnosis, string(icd10JSON), string(symptomsJSON), req.TreatmentPlan,
		req.FollowUpRequired, followUpDate, req.BloodPressureSys, req.BloodPressureDia,
		req.HeartRate, req.Temperature, req.WeightKg, req.HeightCm, req.OxygenSaturation,
		req.Notes, req.IsConfidential,
	)
	if err != nil {
		response.InternalError(c, "failed to create medical record")
		return
	}

	insertedID, _ := result.LastInsertId()
	response.Created(c, "medical record created", gin.H{"id": insertedID})
}
