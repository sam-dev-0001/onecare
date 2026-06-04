package router

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/onecare/backend/internal/config"
	"github.com/onecare/backend/internal/handlers"
	"github.com/onecare/backend/internal/middleware"
)

func Setup(db *sql.DB, cfg *config.Config) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	// Health check — test this first after starting server
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "OneCare API is running"})
	})

	authHandler := handlers.NewAuthHandler(db, cfg)
	doctorHandler := handlers.NewDoctorHandler(db)
	patientHandler := handlers.NewPatientHandler(db)
	appointmentHandler := handlers.NewAppointmentHandler(db)
	historyHandler := handlers.NewMedicalHistoryHandler(db)
	availabilityHandler := handlers.NewAvailabilityHandler(db)
	specializationHandler := handlers.NewSpecializationHandler(db)

	api := r.Group("/api/v1")

	// ── Public routes (NO token needed) ──────────────────────
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/register/doctor", authHandler.RegisterDoctor)
	api.POST("/auth/register/admin", authHandler.RegisterAdmin)
	api.POST("/auth/register/patient", patientHandler.Register)
	api.POST("/auth/forgot-password", authHandler.ForgotPassword)
	api.POST("/auth/reset-password", authHandler.VerifyOTPAndResetPassword)

	// ── Protected routes (token required) ────────────────────
	auth := api.Group("")
	auth.Use(middleware.Auth(cfg))

	// Doctors
	auth.GET("/doctors", doctorHandler.List)
	auth.GET("/doctors/:id", doctorHandler.GetByID)
	auth.PUT("/doctors/:id", doctorHandler.Update)

	// Doctor availability
	auth.GET("/doctors/:id/availability", availabilityHandler.ListForDoctor)
	auth.POST("/doctors/:id/availability", availabilityHandler.Create)
	auth.DELETE("/doctors/:id/availability/:avail_id", availabilityHandler.Deactivate)

	// Admin-only
	admin := api.Group("")
	admin.Use(middleware.Auth(cfg), middleware.RequireRole("admin"))
	admin.POST("/doctors", doctorHandler.Create)
	admin.DELETE("/doctors/:id", doctorHandler.Delete)

	// Patients
	auth.GET("/patients", patientHandler.List)
	auth.GET("/patients/:id", patientHandler.GetByID)
	auth.PUT("/patients/:id", patientHandler.Update)
	auth.PUT("/patients/:id/health-info", patientHandler.UpdateHealthInfo)
	auth.PUT("/patients/:id/emergency-contact", patientHandler.UpdateEmergencyContact)

	// Medical history
	auth.GET("/patients/:id/records", historyHandler.ListForPatient)
	auth.POST("/patients/:id/records", historyHandler.Create)

	// Appointments
	auth.GET("/appointments", appointmentHandler.List)
	auth.GET("/appointments/:id", appointmentHandler.GetByID)
	auth.POST("/appointments", appointmentHandler.Create)
	auth.PATCH("/appointments/:id/status", appointmentHandler.UpdateStatus)

	// Specializations (public)
	api.GET("/specializations", specializationHandler.List)

	return r
}
