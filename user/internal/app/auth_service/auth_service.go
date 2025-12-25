package auth_service

import (
	"errors"
	"time"
	"user/pkg/jwt"
)

type AuthService interface {
	Login(email, password string) (string, error)
	ValidateToken(token string) (*jwt.Claims, error)
	RefreshToken(token string) (string, error)
}

type UserProvider interface {
	ValidateCredentials(email, password string) (*struct {
		ID    int64
		Email string
		Role  string
	}, error)
}

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

func NewAuthService(userProvider UserProvider, secretKey string, tokenDuration time.Duration) AuthService {
	return &authService{
		userProvider: userProvider,
		jwtManager:   jwt.NewJWTManager(secretKey, tokenDuration),
	}
}

func (s *authService) Login(email, password string) (string, error) {

	user, err := s.userProvider.ValidateCredentials(email, password)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

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

	newToken, err := s.jwtManager.GenerateToken(claims.UserID, claims.Email, claims.Role)
	if err != nil {
		return "", errors.New("failed to refresh token")
	}

	return newToken, nil
}
