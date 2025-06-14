package service

import (
	"queue_system/internal/dto/request"
	"queue_system/internal/model"
	"queue_system/internal/repository/mocks"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUserService_CreateUser_Success(t *testing.T) {
	//GIVEN
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	userService := NewUserService(mockUserRepo)

	req := &request.CreateUserRequest{
		Name:  "Test User GoMock",
		Email: "test.gomock@example.com",
		Role:  "member",
	}
	expectedUser := &model.User{
		ID:    1,
		Name:  req.Name,
		Email: req.Email,
		Role:  req.Role,
	}
	mockUserRepo.EXPECT().GetByEmail(req.Email).Return(nil, nil).Times(1)
	mockUserRepo.EXPECT().CreateUser(gomock.Any()).DoAndReturn(
		func(userArg *model.User) (*model.User, error) {
			userArg.ID = 1
			userArg.CreatedAt = time.Now()
			userArg.UpdatedAt = time.Now()
			return userArg, nil
		}).Times(1)

	// WHEN
	createdUser, err := userService.CreateUser(req)

	//THEN
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, expectedUser.Name, createdUser.Name)
	assert.Equal(t, expectedUser.Email, createdUser.Email)
	assert.Equal(t, uint(1), createdUser.ID)
}

func TestUserService_CreateUser_EmailExists(t *testing.T) {
	// GIVEN
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	userService := NewUserService(mockUserRepo)

	request := &request.CreateUserRequest{
		Name:  "Test User GoMock",
		Email: "exists.gomock@example.com",
		Role:  "member",
	}
	exsistingUser := &model.User{
		ID:    1,
		Email: request.Email,
		Name:  "Existing User",
		Role:  request.Role,
	}
	mockUserRepo.EXPECT().GetByEmail(request.Email).Return(exsistingUser, nil).Times(1)
	// WHEN
	createdUser, err := userService.CreateUser(request)

	// THEN
	assert.Error(t, err)
	assert.Nil(t, createdUser)
	assert.Equal(t, ErrEmailExists, err)
}

func TestUserService_GetUserById_Success(t *testing.T) {
	// GIVEN
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepository := mocks.NewMockUserRepository(ctrl)
	userService := NewUserService(mockUserRepository)
	userID := uint(1)

	expectedUser := &model.User{
		ID:    userID,
		Name:  "Found User GoMock",
		Email: "found.gomock@example.com",
		Role:  "admin",
	}

	mockUserRepository.EXPECT().GetById(userID).Return(expectedUser, nil).Times(1)

	//WHEN
	user, err := userService.GetUserById(userID)

	//THEN
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser, user)
}
