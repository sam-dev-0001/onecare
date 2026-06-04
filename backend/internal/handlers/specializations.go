package handlers

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/onecare/backend/internal/models"
	"github.com/onecare/backend/pkg/response"
)

type SpecializationHandler struct {
	db *sql.DB
}

func NewSpecializationHandler(db *sql.DB) *SpecializationHandler {
	return &SpecializationHandler{db: db}
}

// GET /api/v1/specializations
func (h *SpecializationHandler) List(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, name, code, description, parent_id, is_active, created_at
		FROM doctor_specializations
		WHERE is_active = 1
		ORDER BY name`)
	if err != nil {
		response.InternalError(c, "failed to fetch specializations")
		return
	}
	defer rows.Close()

	specs := []models.Specialization{}
	for rows.Next() {
		var s models.Specialization
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Code, &s.Description,
			&s.ParentID, &s.IsActive, &s.CreatedAt,
		); err != nil {
			continue
		}
		specs = append(specs, s)
	}
	response.OK(c, "specializations retrieved", specs)
}
