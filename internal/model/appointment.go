package model

import "time"

type Appointment struct {
	ID            uint      `gorm:"primaryKey"`
	UserID        uint      `gorm:"not null"`
	ParticipantID uint      `gorm:"not null"`
	TimeSlot      time.Time `gorm:"not null"`
	Description   string
	Status        string    `gorm:"default:'pending'"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
