package handlers

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onecare/backend/internal/models"
	"github.com/onecare/backend/pkg/response"
)

type AppointmentHandler struct {
	db *sql.DB
}

func NewAppointmentHandler(db *sql.DB) *AppointmentHandler {
	return &AppointmentHandler{db: db}
}

// POST /api/v1/appointments
func (h *AppointmentHandler) Create(c *gin.Context) {
	var req models.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	patientID, _ := c.Get("user_id")

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		response.BadRequest(c, "invalid scheduled_at format, use RFC3339")
		return
	}
	if scheduledAt.Before(time.Now()) {
		response.BadRequest(c, "scheduled_at must be in the future")
		return
	}

	duration := 30
	if req.DurationMinutes > 0 {
		duration = req.DurationMinutes
	}
	apptType := "in_person"
	if req.AppointmentType != "" {
		apptType = req.AppointmentType
	}

	result, err := h.db.Exec(`
		INSERT INTO appointments
		  (patient_id, doctor_id, scheduled_at, duration_minutes, appointment_type,
		   reason_for_visit, symptoms_description, is_follow_up, parent_appointment_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		patientID, req.DoctorID, scheduledAt, duration, apptType,
		req.ReasonForVisit, req.SymptomsDescription, req.IsFollowUp, req.ParentAppointmentID,
	)
	if err != nil {
		response.Conflict(c, "time slot not available or double-booking detected")
		return
	}

	insertedID, _ := result.LastInsertId()
	response.Created(c, "appointment booked", gin.H{"id": insertedID})
}

// GET /api/v1/appointments  — filtered by role
func (h *AppointmentHandler) List(c *gin.Context) {
	callerID, _ := c.Get("user_id")
	callerRole, _ := c.Get("role")
	status := c.Query("status")

	var query string
	var args []interface{}

	base := `SELECT a.id, a.patient_id, a.doctor_id, a.scheduled_at, a.duration_minutes,
		        a.appointment_type, a.status, a.reason_for_visit, a.is_follow_up,
		        a.created_at, a.updated_at,
		        CONCAT(d.first_name,' ',d.last_name) AS doctor_name,
		        CONCAT(p.first_name,' ',p.last_name) AS patient_name
		FROM appointments a
		JOIN doctors d ON d.id = a.doctor_id
		JOIN patients p ON p.id = a.patient_id`

	switch callerRole.(string) {
	case "patient":
		query = base + " WHERE a.patient_id = ?"
		args = append(args, callerID)
	case "doctor":
		query = base + " WHERE a.doctor_id = ?"
		args = append(args, callerID)
	default: // admin
		query = base + " WHERE 1=1"
	}

	if status != "" {
		query += " AND a.status = ?"
		args = append(args, status)
	}
	query += " ORDER BY a.scheduled_at DESC LIMIT 100"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		response.InternalError(c, "failed to fetch appointments")
		return
	}
	defer rows.Close()

	appts := []models.Appointment{}
	for rows.Next() {
		var a models.Appointment
		if err := rows.Scan(
			&a.ID, &a.PatientID, &a.DoctorID, &a.ScheduledAt, &a.DurationMinutes,
			&a.AppointmentType, &a.Status, &a.ReasonForVisit, &a.IsFollowUp,
			&a.CreatedAt, &a.UpdatedAt, &a.DoctorName, &a.PatientName,
		); err != nil {
			continue
		}
		appts = append(appts, a)
	}
	response.OK(c, "appointments retrieved", appts)
}

// GET /api/v1/appointments/:id
func (h *AppointmentHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid appointment id")
		return
	}

	var a models.Appointment
	err = h.db.QueryRow(`
		SELECT a.id, a.patient_id, a.doctor_id, a.scheduled_at, a.duration_minutes,
		       a.appointment_type, a.status, a.reason_for_visit, a.doctor_notes,
		       a.is_follow_up, a.meeting_url, a.cancellation_reason,
		       a.cancelled_at, a.cancelled_by, a.confirmed_at, a.completed_at,
		       a.created_at, a.updated_at,
		       CONCAT(d.first_name,' ',d.last_name),
		       CONCAT(p.first_name,' ',p.last_name)
		FROM appointments a
		JOIN doctors d ON d.id = a.doctor_id
		JOIN patients p ON p.id = a.patient_id
		WHERE a.id = ?`, id).Scan(
		&a.ID, &a.PatientID, &a.DoctorID, &a.ScheduledAt, &a.DurationMinutes,
		&a.AppointmentType, &a.Status, &a.ReasonForVisit, &a.DoctorNotes,
		&a.IsFollowUp, &a.MeetingURL, &a.CancellationReason,
		&a.CancelledAt, &a.CancelledBy, &a.ConfirmedAt, &a.CompletedAt,
		&a.CreatedAt, &a.UpdatedAt, &a.DoctorName, &a.PatientName,
	)
	if err == sql.ErrNoRows {
		response.NotFound(c, "appointment not found")
		return
	}
	if err != nil {
		response.InternalError(c, "database error")
		return
	}
	response.OK(c, "appointment retrieved", a)
}

// PATCH /api/v1/appointments/:id/status
func (h *AppointmentHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid appointment id")
		return
	}

	var req models.UpdateAppointmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var query string
	var args []interface{}

	switch req.Status {
	case "confirmed":
		query = `UPDATE appointments SET status = 'confirmed', confirmed_at = NOW() WHERE id = ?`
		args = []interface{}{id}
	case "completed":
		query = `UPDATE appointments SET status = 'completed', completed_at = NOW() WHERE id = ?`
		args = []interface{}{id}
	case "cancelled":
		query = `UPDATE appointments SET status = 'cancelled', cancellation_reason = ?, cancelled_at = NOW(), cancelled_by = ? WHERE id = ?`
		args = []interface{}{req.CancellationReason, req.CancelledBy, id}
	default:
		query = `UPDATE appointments SET status = ? WHERE id = ?`
		args = []interface{}{req.Status, id}
	}

	if _, err := h.db.Exec(query, args...); err != nil {
		response.InternalError(c, "failed to update appointment status")
		return
	}
	response.OK(c, "appointment status updated", nil)
}
