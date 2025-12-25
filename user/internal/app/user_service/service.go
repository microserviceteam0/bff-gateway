package user_service

import (
	"errors"
	"regexp"
	"user/internal/domain/entity"
	"user/internal/domain/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(name, email, password, role string) (*entity.User, error)
	GetUserByID(id int64) (*entity.User, error)
	GetUserByEmail(email string) (*entity.User, error)
	GetAllUsers() ([]entity.User, error)
	UpdateUser(id int64, name, email, role string) error
	DeleteUser(id int64) error
	ValidateCredentials(email, password string) (*entity.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(name, email, password, role string) (*entity.User, error) {

	if err := validateUserInput(name, email, password, role); err != nil {
		return nil, err
	}

	existingUser, err := s.repo.FindByEmail(email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Role:     role,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUserByID(id int64) (*entity.User, error) {
	if id == 0 {
		return nil, errors.New("invalid user ID")
	}
	return s.repo.FindByID(id)
}

func (s *userService) GetUserByEmail(email string) (*entity.User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}
	return s.repo.FindByEmail(email)
}

func (s *userService) GetAllUsers() ([]entity.User, error) {
	return s.repo.FindAll()
}

func (s *userService) UpdateUser(id int64, name, email, role string) error {
	if id == 0 {
		return errors.New("invalid user ID")
	}

	user, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	if email != user.Email {
		if err := validateEmail(email); err != nil {
			return err
		}

		existingUser, err := s.repo.FindByEmail(email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return errors.New("email already in use by another user")
		}
		user.Email = email
	}

	if name != "" {
		user.Name = name
	}
	if role != "" {
		user.Role = role
	}

	return s.repo.Update(user)
}

func (s *userService) DeleteUser(id int64) error {
	if id == 0 {
		return errors.New("invalid user ID")
	}
	return s.repo.Delete(id)
}

func (s *userService) ValidateCredentials(email, password string) (*entity.User, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := comparePasswords(user.Password, password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func validateUserInput(name, email, password, role string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if err := validateEmail(email); err != nil {
		return err
	}
	if password == "" {
		return errors.New("password cannot be empty")
	}
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	if role == "" {
		return errors.New("role cannot be empty")
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}

	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func comparePasswords(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
