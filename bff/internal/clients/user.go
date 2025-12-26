package clients

import (
	"context"

	"google.golang.org/grpc"

	userv1 "github.com/microserviceteam0/bff-gateway/bff/api/proto/user"
)

type UserClient interface {
	GetUser(ctx context.Context, id int64, opts ...grpc.CallOption) (*userv1.GetUserResponse, error)
	GetUserByEmail(ctx context.Context, email string, opts ...grpc.CallOption) (*userv1.GetUserResponse, error)
	GetUsers(ctx context.Context, ids []int64, opts ...grpc.CallOption) (*userv1.GetUsersResponse, error)
	UserExists(ctx context.Context, id int64, opts ...grpc.CallOption) (*userv1.UserExistsResponse, error)
	ValidateCredentials(ctx context.Context, email, password string, opts ...grpc.CallOption) (*userv1.ValidateCredentialsResponse, error)
}
