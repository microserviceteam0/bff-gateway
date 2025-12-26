package grpc

import (
	"context"
	"log"
	userv1 "user/api/proto"
	"user/internal/app/user_service"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	userService user_service.UserService
}

func NewServer(userService user_service.UserService) *Server {
	return &Server{
		userService: userService,
	}
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	log.Printf("gRPC GetUser called with ID: %d", req.UserId)

	user, err := s.userService.GetUserByID(req.UserId)
	if err != nil {
		return nil, err
	}

	return &userv1.GetUserResponse{
		User: &userv1.User{
			Id:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}

func (s *Server) GetUserByEmail(ctx context.Context, req *userv1.GetUserByEmailRequest) (*userv1.GetUserResponse, error) {
	log.Printf("gRPC GetUserByEmail called with email: %s", req.Email)

	user, err := s.userService.GetUserByEmail(req.Email)
	if err != nil {
		return nil, err
	}

	return &userv1.GetUserResponse{
		User: &userv1.User{
			Id:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}

func (s *Server) GetUsers(ctx context.Context, req *userv1.GetUsersRequest) (*userv1.GetUsersResponse, error) {
	log.Printf("gRPC GetUsers called with IDs: %v", req.UserIds)

	var users []*userv1.User
	for _, userID := range req.UserIds {
		user, err := s.userService.GetUserByID(userID)
		if err == nil && user != nil {
			users = append(users, &userv1.User{
				Id:    user.ID,
				Name:  user.Name,
				Email: user.Email,
				Role:  user.Role,
			})
		}
	}

	return &userv1.GetUsersResponse{
		Users: users,
	}, nil
}

func (s *Server) UserExists(ctx context.Context, req *userv1.UserExistsRequest) (*userv1.UserExistsResponse, error) {
	log.Printf("gRPC UserExists called with ID: %d", req.UserId)

	user, err := s.userService.GetUserByID(req.UserId)

	return &userv1.UserExistsResponse{
		Exists: user != nil && err == nil,
	}, nil
}

func (s *Server) ValidateCredentials(ctx context.Context, req *userv1.ValidateCredentialsRequest) (*userv1.ValidateCredentialsResponse, error) {
	log.Printf("gRPC ValidateCredentials called for email: %s", req.Email)

	user, err := s.userService.ValidateCredentials(req.Email, req.Password)
	if err != nil {
		return &userv1.ValidateCredentialsResponse{
			Valid: false,
		}, nil
	}

	return &userv1.ValidateCredentialsResponse{
		Valid: true,
		User: &userv1.User{
			Id:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}
