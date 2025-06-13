package service

import (
	"errors"
	"queue_system/internal/dto/request"
	"queue_system/internal/model"
	"queue_system/internal/repository"

	"github.com/rs/zerolog/log"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailExists      = errors.New("email already exists")
	ErrUpdateFailed     = errors.New("failed to update user")
	ErrCreateUserFailed = errors.New("failed to create user")
)

type UserService interface {
	CreateUser(req *request.CreateUserRequest) (*model.User, error)
	GetUserById(id uint) (*model.User, error)
}

type userService struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userService{
		userRepository: userRepository,
	}
}

func (us *userService) CreateUser(req *request.CreateUserRequest) (*model.User, error) {
	existingUser, err := us.userRepository.GetByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Error checking existing email")
		return nil, err
	}
	if existingUser != nil {
		log.Error().Err(ErrEmailExists).Msg("Email already exists")
		return nil, ErrEmailExists
	}
	user := &model.User{
		Name:  req.Name,
		Email: req.Email,
		Role:  req.Role,
	}
	createdUser, err := us.userRepository.CreateUser(user)
	if err != nil {
		log.Error().Err(err).Msg("Error creating user")
		return nil, ErrCreateUserFailed
	}
	return createdUser, nil
}

func (us *userService) GetUserById(id uint) (*model.User, error) {
	user, err := us.userRepository.GetById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(ErrUserNotFound).Msg("User not found")
			return nil, ErrUserNotFound
		}
		log.Error().Err(err).Msg("Error getting user")
		return nil, err
	}
	return user, nil
}
