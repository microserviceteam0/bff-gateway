package service

import (
	"context"

	"github.com/microserviceteam0/bff-gateway/bff/internal/dto"
)

func (s *bffService) Register(ctx context.Context, req dto.RegisterUserRequestDTO) (*dto.UserResponseDTO, error) {
	user, err := s.userHTTPClient.CreateUser(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &dto.UserResponseDTO{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *bffService) Login(ctx context.Context, req dto.LoginRequestDTO) (*dto.LoginResponseDTO, error) {
	resp, err := s.authClient.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponseDTO{Token: resp.Token}, nil
}

func (s *bffService) GetUserProfile(ctx context.Context, userID int64, userRole string) (*dto.UserResponseDTO, error) {
	ctx = withAuthMetadata(ctx, userID, userRole)
	resp, err := s.userClient.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	user := resp.GetUser()
	return &dto.UserResponseDTO{
		ID:    user.GetId(),
		Name:  user.GetName(),
		Email: user.GetEmail(),
	}, nil
}
