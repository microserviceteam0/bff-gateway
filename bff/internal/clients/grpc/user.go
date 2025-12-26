package grpc

import (
	"context"

	"github.com/microserviceteam0/bff-gateway/bff/internal/clients"
	"google.golang.org/grpc"
	
	userv1 "github.com/microserviceteam0/bff-gateway/bff/api/proto/user"
)

type userClient struct {
	api userv1.UserServiceClient
}

func NewUserClient(conn *grpc.ClientConn) clients.UserClient {
	return &userClient{
		api: userv1.NewUserServiceClient(conn),
	}
}

func (c *userClient) GetUser(ctx context.Context, id int64, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
	resp, err := c.api.GetUser(ctx, &userv1.GetUserRequest{UserId: id}, opts...)
	return resp, clients.MapGRPCError(err)
}

func (c *userClient) GetUserByEmail(ctx context.Context, email string, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
	resp, err := c.api.GetUserByEmail(ctx, &userv1.GetUserByEmailRequest{Email: email}, opts...)
	return resp, clients.MapGRPCError(err)
}

func (c *userClient) GetUsers(ctx context.Context, ids []int64, opts ...grpc.CallOption) (*userv1.GetUsersResponse, error) {
	resp, err := c.api.GetUsers(ctx, &userv1.GetUsersRequest{UserIds: ids}, opts...)
	return resp, clients.MapGRPCError(err)
}

func (c *userClient) UserExists(ctx context.Context, id int64, opts ...grpc.CallOption) (*userv1.UserExistsResponse, error) {
	resp, err := c.api.UserExists(ctx, &userv1.UserExistsRequest{UserId: id}, opts...)
	return resp, clients.MapGRPCError(err)
}

func (c *userClient) ValidateCredentials(ctx context.Context, email, password string, opts ...grpc.CallOption) (*userv1.ValidateCredentialsResponse, error) {
	resp, err := c.api.ValidateCredentials(ctx, &userv1.ValidateCredentialsRequest{
		Email:    email,
		Password: password,
	}, opts...)
	return resp, clients.MapGRPCError(err)
}