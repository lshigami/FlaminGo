package service

import (
	"errors"
	"queue_system/internal/dto/request"
	"queue_system/internal/enums"
	"queue_system/internal/model"
	"queue_system/internal/repository"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var (
	ErrAppointmentNotFound       = errors.New("appointment not found")
	ErrInvalidTimeFormat         = errors.New("invalid time format, use RFC3339 (e.g., 2024-01-01T10:00:00Z)")
	ErrEndTimeBeforeStartTime    = errors.New("end time must be after start time")
	ErrAppointmentConflict       = errors.New("time slot conflicts with an existing appointment for one of the participants")
	ErrUserOrParticipantNotFound = errors.New("creator (user) or participant not found")
	ErrCannotBookWithSelf        = errors.New("user cannot book an appointment with themselves")
	ErrInvalidAppointmentStatus  = errors.New("invalid appointment status")
	ErrCreateAppointmentFailed   = errors.New("failed to create appointment")
	ErrUpdateAppointmentFailed   = errors.New("failed to update appointment")
)

type AppointmentService interface {
	CreateAppointment(req *request.AppointmentRequest) (*model.Appointment, error)
	GetAppointmentByID(id uint) (*model.Appointment, error)
}

type appointmentService struct {
	appointmentRepository repository.AppointmentRepository
	userRepository        repository.UserRepository
	db                    *gorm.DB
}

func NewAppointmentService(appointmentRepository repository.AppointmentRepository, userRepository repository.UserRepository, db *gorm.DB) AppointmentService {
	return &appointmentService{
		appointmentRepository: appointmentRepository,
		userRepository:        userRepository,
		db:                    db,
	}
}

func (as *appointmentService) CreateAppointment(req *request.AppointmentRequest) (*model.Appointment, error) {
	if req.UserID == req.ParticipantID {
		return nil, ErrCannotBookWithSelf
	}
	start_time, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		log.Warn().Err(err).Str("startTime", req.StartTime).Msg("Failed to parse start time")
		return nil, ErrInvalidTimeFormat
	}
	end_time, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		log.Warn().Err(err).Str("endTime", req.EndTime).Msg("Failed to parse end time")
		return nil, ErrInvalidTimeFormat
	}
	if end_time.Before(start_time) {
		return nil, ErrEndTimeBeforeStartTime
	}

	appointment := &model.Appointment{
		UserID:        req.UserID,
		ParticipantID: req.ParticipantID,
		StartTime:     start_time,
		EndTime:       end_time,
		Description:   req.Description,
		Status:        string(enums.Pending),
	}
	tx := as.db.Begin()

	user, err := as.userRepository.GetById(req.UserID)
	if err != nil && user == nil {
		tx.Rollback()
		return nil, ErrUserOrParticipantNotFound
	}

	participant, err := as.userRepository.GetById(req.ParticipantID)
	if err != nil && participant == nil {
		tx.Rollback()
		return nil, ErrUserOrParticipantNotFound
	}
	conflictingAppointments, err := as.appointmentRepository.FindConflictingAppointments(tx, appointment)
	if err != nil {
		log.Error().Err(err).Msg("Error checking for conflicting appointments")
		return nil, err
	}
	if len(conflictingAppointments) > 0 {
		tx.Rollback()
		log.Warn().Msg("Conflicting appointments found")
		return nil, ErrAppointmentConflict
	}

	if err := as.appointmentRepository.CreateWithTx(tx, appointment); err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("Error creating appointment")
		return nil, ErrCreateAppointmentFailed
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("Error committing transaction")
		return nil, ErrUpdateAppointmentFailed
	}
	return appointment, nil

}

func (as *appointmentService) GetAppointmentByID(id uint) (*model.Appointment, error) {
	appointment, err := as.appointmentRepository.GetByID(id)
	if err != nil {
		log.Error().Err(err).Uint("appointmentID", id).Msg("Error fetching appointment by ID")
		return nil, err
	}
	if appointment == nil {
		return nil, ErrAppointmentNotFound
	}
	return appointment, nil
}
