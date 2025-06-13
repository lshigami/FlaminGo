package controller

import (
	"net/http"
	"queue_system/internal/dto/request"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type AppointmentController struct {
	appointmentService service.AppointmentService
}

func NewAppointmentController(appointmentService service.AppointmentService) *AppointmentController {
	return &AppointmentController{
		appointmentService: appointmentService,
	}
}

func (c *AppointmentController) CreateAppointment(ctx *gin.Context) {
	var req request.AppointmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("Failed to bind appointment")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdAppointment, err := c.appointmentService.CreateAppointment(ctx, &req)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create appointment")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, createdAppointment)

}
