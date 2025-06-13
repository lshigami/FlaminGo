package controller

import (
	"errors"
	"net/http"
	"queue_system/internal/dto/request"
	"queue_system/internal/service"
	"strconv"

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

	createdAppointment, err := c.appointmentService.CreateAppointment(&req)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create appointment")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, createdAppointment)

}
func (c *AppointmentController) GetAppointmentByID(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID format"})
		return
	}

	appointment, err := c.appointmentService.GetAppointmentByID(uint(id))
	if err != nil {
		if errors.Is(err, service.ErrAppointmentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		log.Error().Err(err).Msg("Failed to get appointment by ID")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve appointment"})
		return
	}
	ctx.JSON(http.StatusOK, appointment)
}
