// cmd_auth/internal/app/auth_service/auth_service.go
package auth_service

import (
	"errors"
	"time"
	"user/pkg/jwt"
)

// AuthService интерфейс для аутентификации
type AuthService interface {
	Login(email, password string) (string, error)
	ValidateToken(token string) (*jwt.Claims, error)
	RefreshToken(token string) (string, error)
}

// UserProvider интерфейс для получения данных пользователя
// Это адаптер к основному UserService
type UserProvider interface {
	ValidateCredentials(email, password string) (*struct {
		ID    int64
		Email string
		Role  string
	}, error)
}

// Adapter для основного UserService
type userServiceAdapter struct {
	userService interface {
		ValidateCredentials(email, password string) (*struct {
			ID    int64
			Name  string
			Email string
			Role  string
		}, error)
	}
}

func NewUserServiceAdapter(userService interface {
	ValidateCredentials(email, password string) (*struct {
		ID    int64
		Name  string
		Email string
		Role  string
	}, error)
}) UserProvider {
	return &userServiceAdapter{userService: userService}
}

func (a *userServiceAdapter) ValidateCredentials(email, password string) (*struct {
	ID    int64
	Email string
	Role  string
}, error) {
	user, err := a.userService.ValidateCredentials(email, password)
	if err != nil {
		return nil, err
	}

	return &struct {
		ID    int64
		Email string
		Role  string
	}{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

type authService struct {
	userProvider UserProvider
	jwtManager   *jwt.JWTManager
}

// NewAuthService создаёт новый экземпляр AuthService
func NewAuthService(userProvider UserProvider, secretKey string, tokenDuration time.Duration) AuthService {
	return &authService{
		userProvider: userProvider,
		jwtManager:   jwt.NewJWTManager(secretKey, tokenDuration),
	}
}

func (s *authService) Login(email, password string) (string, error) {
	// Используем UserProvider для проверки учетных данных
	user, err := s.userProvider.ValidateCredentials(email, password)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// Генерация JWT токена
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}

func (s *authService) ValidateToken(token string) (*jwt.Claims, error) {
	return s.jwtManager.ValidateToken(token)
}

func (s *authService) RefreshToken(token string) (string, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return "", err
	}

	// Генерируем новый токен с теми же данными
	newToken, err := s.jwtManager.GenerateToken(claims.UserID, claims.Email, claims.Role)
	if err != nil {
		return "", errors.New("failed to refresh token")
	}

	return newToken, nil
}
