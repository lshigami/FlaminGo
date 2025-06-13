package repository

import (
	"queue_system/internal/dto/request"
	"queue_system/internal/model"
)

type AppointmentRepository interface {
	CreateAppointment(req request.AppointmentRequest) (model.Appointment, error)
}
