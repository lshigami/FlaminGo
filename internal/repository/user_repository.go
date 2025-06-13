package repository

import (
	"errors"
	"queue_system/internal/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *model.User) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetById(id uint) (*model.User, error)
	UpdateUser(user *model.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (ur *userRepository) CreateUser(user *model.User) (*model.User, error) {
	return user, ur.db.Create(user).Error
}

func (ur *userRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	if err := ur.db.Where("email=?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (ur *userRepository) GetById(id uint) (*model.User, error) {
	var user model.User
	if err := ur.db.Where("id=?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (ur *userRepository) UpdateUser(user *model.User) error {
	return ur.db.Save(user).Error
}
