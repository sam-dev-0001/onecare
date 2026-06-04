package models

import "time"

// ─── Doctor ──────────────────────────────────────────────────

type Doctor struct {
	ID                int        `json:"id"`
	FirstName         string     `json:"first_name"`
	LastName          string     `json:"last_name"`
	Email             string     `json:"email"`
	Phone             *string    `json:"phone,omitempty"`
	Gender            *string    `json:"gender,omitempty"`
	DateOfBirth       *time.Time `json:"date_of_birth,omitempty"`
	ProfilePhotoURL   *string    `json:"profile_photo_url,omitempty"`
	MedicalLicenseNo  string     `json:"medical_license_no"`
	YearsOfExperience *int       `json:"years_of_experience,omitempty"`
	Bio               *string    `json:"bio,omitempty"`
	ConsultationFee   *float64   `json:"consultation_fee,omitempty"`
	Rating            *float64   `json:"rating,omitempty"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CreateDoctorRequest struct {
	FirstName         string   `json:"first_name"        binding:"required,min=2,max=80"`
	LastName          string   `json:"last_name"         binding:"required,min=2,max=80"`
	Email             string   `json:"email"             binding:"required,email"`
	Password          string   `json:"password"          binding:"required,min=8"`
	Phone             *string  `json:"phone"`
	Gender            *string  `json:"gender"            binding:"omitempty,oneof=male female other prefer_not_to_say"`
	MedicalLicenseNo  string   `json:"medical_license_no" binding:"required"`
	YearsOfExperience *int     `json:"years_of_experience"`
	Bio               *string  `json:"bio"`
	ConsultationFee   *float64 `json:"consultation_fee"`
	SpecializationIDs []int    `json:"specialization_ids"`
}

// ─── Patient ─────────────────────────────────────────────────

type Patient struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	// FirstName             string     `json:"first_name"`
	// LastName              string     `json:"last_name"`
	Email                 *string    `json:"email,omitempty"`
	Phone                 string     `json:"phone"`
	Gender                *string    `json:"gender,omitempty"`
	DateOfBirth           *time.Time `json:"date_of_birth,omitempty"`
	BloodGroup            *string    `json:"blood_group,omitempty"`
	NationalID            *string    `json:"national_id,omitempty"`
	ProfilePhotoURL       *string    `json:"profile_photo_url,omitempty"`
	EmergencyContactName  *string    `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone *string    `json:"emergency_contact_phone,omitempty"`
	EmergencyContactRel   *string    `json:"emergency_contact_rel,omitempty"`
	AddressLine1          *string    `json:"address_line1,omitempty"`
	AddressLine2          *string    `json:"address_line2,omitempty"`
	City                  *string    `json:"city,omitempty"`
	State                 *string    `json:"state,omitempty"`
	PostalCode            *string    `json:"postal_code,omitempty"`
	Country               string     `json:"country"`
	KnownAllergies        []string   `json:"known_allergies,omitempty"`
	ChronicConditions     []string   `json:"chronic_conditions,omitempty"`
	IsActive              bool       `json:"is_active"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type CreatePatientRequest struct {
	Name string `json:"name"  binding:"required, min=2,max=80"`
	// FirstName  string  `json:"first_name" binding:"required,min=2,max=80"`
	// LastName   string  `json:"last_name"  binding:"required,min=2,max=80"`
	Email      *string `json:"email"      binding:"omitempty,email"`
	Password   string  `json:"password"   binding:"required,min=8"`
	Phone      string  `json:"phone"      binding:"required"`
	Gender     *string `json:"gender"     binding:"omitempty,oneof=male female other prefer_not_to_say"`
	BloodGroup *string `json:"blood_group"`
	Country    string  `json:"country"`
}

// ─── Appointment ─────────────────────────────────────────────

type Appointment struct {
	ID                  int        `json:"id"`
	PatientID           int        `json:"patient_id"`
	DoctorID            int        `json:"doctor_id"`
	ScheduledAt         time.Time  `json:"scheduled_at"`
	DurationMinutes     int        `json:"duration_minutes"`
	AppointmentType     string     `json:"appointment_type"`
	Status              string     `json:"status"`
	ReasonForVisit      *string    `json:"reason_for_visit,omitempty"`
	SymptomsDescription *string    `json:"symptoms_description,omitempty"`
	DoctorNotes         *string    `json:"doctor_notes,omitempty"`
	IsFollowUp          bool       `json:"is_follow_up"`
	ParentAppointmentID *int       `json:"parent_appointment_id,omitempty"`
	MeetingURL          *string    `json:"meeting_url,omitempty"`
	CancellationReason  *string    `json:"cancellation_reason,omitempty"`
	CancelledAt         *time.Time `json:"cancelled_at,omitempty"`
	CancelledBy         *string    `json:"cancelled_by,omitempty"`
	ConfirmedAt         *time.Time `json:"confirmed_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	// Joined fields
	DoctorName  *string `json:"doctor_name,omitempty"`
	PatientName *string `json:"patient_name,omitempty"`
}

type CreateAppointmentRequest struct {
	DoctorID            int     `json:"doctor_id"         binding:"required"`
	ScheduledAt         string  `json:"scheduled_at"      binding:"required"` // RFC3339
	DurationMinutes     int     `json:"duration_minutes"`
	AppointmentType     string  `json:"appointment_type"  binding:"omitempty,oneof=in_person teleconsult home_visit"`
	ReasonForVisit      *string `json:"reason_for_visit"`
	SymptomsDescription *string `json:"symptoms_description"`
	IsFollowUp          bool    `json:"is_follow_up"`
	ParentAppointmentID *int    `json:"parent_appointment_id"`
}

type UpdateAppointmentStatusRequest struct {
	Status             string  `json:"status"              binding:"required,oneof=confirmed cancelled completed no_show"`
	CancellationReason *string `json:"cancellation_reason"`
	CancelledBy        *string `json:"cancelled_by"        binding:"omitempty,oneof=patient doctor admin"`
}

// ─── Medical History ─────────────────────────────────────────

type MedicalHistory struct {
	ID               int        `json:"id"`
	PatientID        int        `json:"patient_id"`
	DoctorID         *int       `json:"doctor_id,omitempty"`
	AppointmentID    *int       `json:"appointment_id,omitempty"`
	RecordType       string     `json:"record_type"`
	RecordDate       time.Time  `json:"record_date"`
	ChiefComplaint   *string    `json:"chief_complaint,omitempty"`
	Diagnosis        *string    `json:"diagnosis,omitempty"`
	ICD10Codes       []string   `json:"icd10_codes,omitempty"`
	Symptoms         []string   `json:"symptoms,omitempty"`
	TreatmentPlan    *string    `json:"treatment_plan,omitempty"`
	FollowUpRequired bool       `json:"follow_up_required"`
	FollowUpDate     *time.Time `json:"follow_up_date,omitempty"`
	BloodPressureSys *int       `json:"blood_pressure_sys,omitempty"`
	BloodPressureDia *int       `json:"blood_pressure_dia,omitempty"`
	HeartRate        *int       `json:"heart_rate,omitempty"`
	Temperature      *float64   `json:"temperature,omitempty"`
	WeightKg         *float64   `json:"weight_kg,omitempty"`
	HeightCm         *float64   `json:"height_cm,omitempty"`
	OxygenSaturation *float64   `json:"oxygen_saturation,omitempty"`
	BMI              *float64   `json:"bmi,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
	IsConfidential   bool       `json:"is_confidential"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CreateMedicalHistoryRequest struct {
	PatientID        int      `json:"patient_id"   binding:"required"`
	AppointmentID    *int     `json:"appointment_id"`
	RecordType       string   `json:"record_type"  binding:"required,oneof=consultation lab_result prescription imaging procedure vaccination discharge_summary referral"`
	RecordDate       string   `json:"record_date"`
	ChiefComplaint   *string  `json:"chief_complaint"`
	Diagnosis        *string  `json:"diagnosis"`
	ICD10Codes       []string `json:"icd10_codes"`
	Symptoms         []string `json:"symptoms"`
	TreatmentPlan    *string  `json:"treatment_plan"`
	FollowUpRequired bool     `json:"follow_up_required"`
	FollowUpDate     *string  `json:"follow_up_date"`
	BloodPressureSys *int     `json:"blood_pressure_sys"`
	BloodPressureDia *int     `json:"blood_pressure_dia"`
	HeartRate        *int     `json:"heart_rate"`
	Temperature      *float64 `json:"temperature"`
	WeightKg         *float64 `json:"weight_kg"`
	HeightCm         *float64 `json:"height_cm"`
	OxygenSaturation *float64 `json:"oxygen_saturation"`
	Notes            *string  `json:"notes"`
	IsConfidential   bool     `json:"is_confidential"`
}

// ─── Specialization ──────────────────────────────────────────

type Specialization struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description *string   `json:"description,omitempty"`
	ParentID    *int      `json:"parent_id,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// ─── Auth ────────────────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"     binding:"required,oneof=doctor patient admin"`
}

type LoginResponse struct {
	Token   string      `json:"token"`
	Expires int64       `json:"expires_at"`
	User    interface{} `json:"user"`
}

// ─── Availability ────────────────────────────────────────────

type Availability struct {
	ID           int        `json:"id"`
	DoctorID     int        `json:"doctor_id"`
	DayOfWeek    string     `json:"day_of_week"`
	StartTime    string     `json:"start_time"`
	EndTime      string     `json:"end_time"`
	SlotDuration int        `json:"slot_duration"`
	MaxSlots     *int       `json:"max_slots,omitempty"`
	Location     *string    `json:"location,omitempty"`
	IsActive     bool       `json:"is_active"`
	ValidFrom    time.Time  `json:"valid_from"`
	ValidUntil   *time.Time `json:"valid_until,omitempty"`
}

type CreateAvailabilityRequest struct {
	DayOfWeek    string  `json:"day_of_week"   binding:"required,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	StartTime    string  `json:"start_time"    binding:"required"` // HH:MM
	EndTime      string  `json:"end_time"      binding:"required"`
	SlotDuration int     `json:"slot_duration" binding:"required,min=5,max=120"`
	MaxSlots     *int    `json:"max_slots"`
	Location     *string `json:"location"`
}

// ─── Password Reset ──────────────────────────────────────────

type ForgotPasswordRequest struct {
	Email    string `json:"email"     binding:"required,email"`
	UserType string `json:"user_type" binding:"required,oneof=doctor patient admin"`
}

type ForgotPasswordResponse struct {
	ResetToken string `json:"reset_token"`
}

type ResetPasswordRequest struct {
	ResetToken  string `json:"reset_token" binding:"required"`
	OTP         string `json:"otp"         binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ─── Patient Update ──────────────────────────────────────────

type UpdatePatientRequest struct {
	FirstName       *string `json:"first_name"`
	LastName        *string `json:"last_name"`
	Gender          *string `json:"gender"                 binding:"omitempty,oneof=male female other prefer_not_to_say"`
	DateOfBirth     *string `json:"date_of_birth"`
	BloodGroup      *string `json:"blood_group"`
	NationalID      *string `json:"national_id"`
	ProfilePhotoURL *string `json:"profile_photo_url"`
	AddressLine1    *string `json:"address_line1"`
	AddressLine2    *string `json:"address_line2"`
	City            *string `json:"city"`
	State           *string `json:"state"`
	PostalCode      *string `json:"postal_code"`
	Country         *string `json:"country"`
}

type UpdateHealthInfoRequest struct {
	KnownAllergies    []string `json:"known_allergies"`
	ChronicConditions []string `json:"chronic_conditions"`
}

type UpdateEmergencyContactRequest struct {
	EmergencyContactName  string `json:"emergency_contact_name"  binding:"required,min=2"`
	EmergencyContactPhone string `json:"emergency_contact_phone" binding:"required"`
	EmergencyContactRel   string `json:"emergency_contact_rel"   binding:"required,min=2"`
}
