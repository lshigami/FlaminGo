package request

type AppointmentRequest struct {
	UserID        uint   `json:"user_id" binding:"required"`
	ParticipantID uint   `json:"participant_id" binding:"required"`
	TimeSlot      string `json:"time_slot" binding:"required"`
	Description   string `json:"description"`
}
