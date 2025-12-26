package service

import (
	"context"
	"strconv"

	"github.com/microserviceteam0/bff-gateway/bff/internal/clients"
	"github.com/microserviceteam0/bff-gateway/bff/internal/dto"
	"google.golang.org/grpc/metadata"
)

type BFFService interface {
	GetOrderDetails(ctx context.Context, userID int64, userRole string, orderID int64) (*dto.OrderResponseDTO, error)
	
	Register(ctx context.Context, req dto.RegisterUserRequestDTO) (*dto.UserResponseDTO, error)
	Login(ctx context.Context, req dto.LoginRequestDTO) (*dto.LoginResponseDTO, error)
	CreateOrder(ctx context.Context, userID int64, userRole string, req dto.CreateOrderRequestDTO) (*dto.OrderResponseDTO, error)
	CancelOrder(ctx context.Context, userID int64, userRole string, orderID int64, reason string) error
	GetUserProfile(ctx context.Context, userID int64, userRole string) (*dto.UserResponseDTO, error)
	ListProducts(ctx context.Context) ([]*dto.ProductResponseDTO, error)
}

type bffService struct {
	userClient        clients.UserClient
	orderClient       clients.OrderClient
	productClient     clients.ProductClient
	authClient        clients.AuthClient
	userHTTPClient    clients.UserHTTPClient
	productHTTPClient clients.ProductHTTPClient
}

func NewBFFService(
	userClient clients.UserClient,
	orderClient clients.OrderClient,
	productClient clients.ProductClient,
	authClient clients.AuthClient,
	userHTTPClient clients.UserHTTPClient,
	productHTTPClient clients.ProductHTTPClient,
) BFFService {
	return &bffService{
		userClient:        userClient,
		orderClient:       orderClient,
		productClient:     productClient,
		authClient:        authClient,
		userHTTPClient:    userHTTPClient,
		productHTTPClient: productHTTPClient,
	}
}

func withAuthMetadata(ctx context.Context, userID int64, role string) context.Context {
	md := metadata.Pairs(
		"x-user-id", strconv.FormatInt(userID, 10),
		"x-user-role", role, 
	)
	return metadata.NewOutgoingContext(ctx, md)
}