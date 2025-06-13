package repository

import (
	"errors"
	"queue_system/internal/model"

	"gorm.io/gorm"
)

type AppointmentRepository interface {
	CreateWithTx(tx *gorm.DB, appointment *model.Appointment) error
	GetByID(id uint) (*model.Appointment, error)
	FindConflictingAppointments(tx *gorm.DB, appointment *model.Appointment) ([]model.Appointment, error)
}

type appointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepository{db: db}
}

func (ar *appointmentRepository) CreateWithTx(tx *gorm.DB, appointment *model.Appointment) error {
	return tx.Create(appointment).Error
}

func (ar *appointmentRepository) GetByID(id uint) (*model.Appointment, error) {
	var appointment model.Appointment

	if err := ar.db.Where("id = ?", id).First(&appointment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &appointment, nil

}

func (ar *appointmentRepository) FindConflictingAppointments(tx *gorm.DB, req *model.Appointment) ([]model.Appointment, error) {

	var conflictingAppointments []model.Appointment

	query := tx.Model(&model.Appointment{}).
		Where(tx.Where("start_time<? AND end_time>?", req.EndTime, req.StartTime)).
		Where(tx.Where("user_id=?", req.UserID).Or("participant_id=?", req.ParticipantID)).
		Where(tx.Where("status NOT IN (?)", []string{"cancelled", "completed"}))

	if err := query.Find(&conflictingAppointments).Error; err != nil {
		return nil, err
	}
	if len(conflictingAppointments) > 0 {
		return conflictingAppointments, nil
	}
	return nil, nil
}
