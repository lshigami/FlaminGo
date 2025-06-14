package controller

import (
	"errors"
	"net/http"
	"queue_system/internal/service"
	"strconv"

	"queue_system/internal/dto/request"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type UserController struct {
	UserService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		UserService: userService,
	}
}

func (uc *UserController) CreateUser(c *gin.Context) {
	var req request.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userResponse, err := uc.UserService.CreateUser(&req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("CreateUser: Service error") // Log lỗi từ service
		if errors.Is(err, service.ErrEmailExists) {                                     // KIỂM TRA LỖI CỤ THỂ
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // TRẢ VỀ 409
			return
		}
		if errors.Is(err, service.ErrCreateUserFailed) { // Xử lý lỗi chung hơn nếu có
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Lỗi không xác định khác
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred while creating the user."})
		return
	}

	c.JSON(http.StatusCreated, userResponse)
}

func (uc *UserController) GetUserById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	user, err := uc.UserService.GetUserById(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}
