package request

type AppointmentRequest struct {
	UserID        uint   `json:"user_id" binding:"required"`
	ParticipantID uint   `json:"participant_id" binding:"required"`
	StartTime     string `json:"start_time" binding:"required"`
	EndTime       string `json:"end_time" binding:"required"`
	Description   string `json:"description"`
}

type UpdateAppointmentRequest struct {
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
	Description *string `json:"description"`
	Status      *string `json:"status" binding:"omitempty,oneof=pending confirmed cancelled"`
}
