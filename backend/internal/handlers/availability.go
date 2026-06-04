package handlers

import (
	"database/sql"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/onecare/backend/internal/models"
	"github.com/onecare/backend/pkg/response"
)

type AvailabilityHandler struct {
	db *sql.DB
}

func NewAvailabilityHandler(db *sql.DB) *AvailabilityHandler {
	return &AvailabilityHandler{db: db}
}

// GET /api/v1/doctors/:id/availability
func (h *AvailabilityHandler) ListForDoctor(c *gin.Context) {
	doctorID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid doctor id")
		return
	}

	rows, err := h.db.Query(`
		SELECT id, doctor_id, day_of_week, start_time, end_time,
		       slot_duration, max_slots, location, is_active, valid_from, valid_until
		FROM doctor_availability
		WHERE doctor_id = ? AND is_active = 1
		ORDER BY FIELD(day_of_week, 'monday','tuesday','wednesday','thursday','friday','saturday','sunday')`,
		doctorID,
	)
	if err != nil {
		response.InternalError(c, "failed to fetch availability")
		return
	}
	defer rows.Close()

	slots := []models.Availability{}
	for rows.Next() {
		var a models.Availability
		if err := rows.Scan(
			&a.ID, &a.DoctorID, &a.DayOfWeek, &a.StartTime, &a.EndTime,
			&a.SlotDuration, &a.MaxSlots, &a.Location, &a.IsActive, &a.ValidFrom, &a.ValidUntil,
		); err != nil {
			continue
		}
		slots = append(slots, a)
	}
	response.OK(c, "availability retrieved", slots)
}

// POST /api/v1/doctors/:id/availability  (doctor or admin)
func (h *AvailabilityHandler) Create(c *gin.Context) {
	doctorID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid doctor id")
		return
	}

	callerID, _ := c.Get("user_id")
	callerRole, _ := c.Get("role")
	if callerRole.(string) != "admin" && callerID.(int) != doctorID {
		response.Unauthorized(c, "cannot modify another doctor's availability")
		return
	}

	var req models.CreateAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.db.Exec(`
		INSERT INTO doctor_availability
		  (doctor_id, day_of_week, start_time, end_time, slot_duration, max_slots, location)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		doctorID, req.DayOfWeek, req.StartTime, req.EndTime,
		req.SlotDuration, req.MaxSlots, req.Location,
	)
	if err != nil {
		response.Conflict(c, "active availability for this day already exists")
		return
	}

	insertedID, _ := result.LastInsertId()
	response.Created(c, "availability created", gin.H{"id": insertedID})
}

// DELETE /api/v1/doctors/:id/availability/:avail_id
func (h *AvailabilityHandler) Deactivate(c *gin.Context) {
	availID, err := strconv.Atoi(c.Param("avail_id"))
	if err != nil {
		response.BadRequest(c, "invalid availability id")
		return
	}

	_, err = h.db.Exec(`UPDATE doctor_availability SET is_active = 0 WHERE id = ?`, availID)
	if err != nil {
		response.InternalError(c, "failed to deactivate availability")
		return
	}
	response.OK(c, "availability deactivated", nil)
}
