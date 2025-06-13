package service

import (
	"queue_system/internal/dto/request"
	"queue_system/internal/model"
	"queue_system/internal/repository"
)

type AppointmentService interface {
	CreateAppointmentService(req request.AppointmentRequest) (model.Appointment, error)
}

type appointmentService struct {
	appointmentRepository repository.AppointmentRepository
}

func NewAppointmentService(appointmentRepository repository.AppointmentRepository) AppointmentService {
	return &appointmentService{
		appointmentRepository: appointmentRepository,
	}
}

func (as *appointmentService) CreateAppointmentService(req request.AppointmentRequest) (model.Appointment, error) {

}
