package repository

import (
	"gorm.io/gorm"
	"user/internal/domain/entity"
)

type UserRepository interface {
	Create(user *entity.User) error
	FindByID(id int64) (*entity.User, error) // изменил uint → int64
	FindByEmail(email string) (*entity.User, error)
	Update(user *entity.User) error
	Delete(id int64) error // изменил uint → int64
	FindAll() ([]entity.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id int64) (*entity.User, error) { // int64
	var user entity.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *userRepository) FindByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepository) Update(user *entity.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id int64) error { // int64
	return r.db.Delete(&entity.User{}, id).Error
}

func (r *userRepository) FindAll() ([]entity.User, error) {
	var users []entity.User
	err := r.db.Find(&users).Error
	return users, err
}
