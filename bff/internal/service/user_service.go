package service

import (
	"context"
	"fmt"

	orderv1 "github.com/microserviceteam0/bff-gateway/bff/api/proto/order/v1"
	productv1 "github.com/microserviceteam0/bff-gateway/bff/api/proto/product"
	userv1 "github.com/microserviceteam0/bff-gateway/bff/api/proto/user"
	"github.com/microserviceteam0/bff-gateway/bff/internal/dto"
	"golang.org/x/sync/errgroup"
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

func (s *bffService) GetUserProfile(ctx context.Context, userID int64, userRole string) (*dto.UserProfileDTO, error) {
	ctx = withAuthMetadata(ctx, userID, userRole)

	var (
		userResp   *userv1.GetUserResponse
		ordersResp *orderv1.GetUserOrdersResponse
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		userResp, err = s.userClient.GetUser(gCtx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		ordersResp, err = s.orderClient.GetUserOrders(gCtx, &orderv1.GetUserOrdersRequest{UserId: userID})
		if err != nil {
			return fmt.Errorf("failed to get orders: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	user := userResp.GetUser()
	orders := ordersResp.GetOrders()

	productIDsMap := make(map[int64]struct{})
	for _, order := range orders {
		for _, item := range order.GetItems() {
			productIDsMap[item.GetProductId()] = struct{}{}
		}
	}

	productIDs := make([]int64, 0, len(productIDsMap))
	for id := range productIDsMap {
		productIDs = append(productIDs, id)
	}

	var productsResp *productv1.ProductsResponse
	if len(productIDs) > 0 {
		var err error
		productsResp, err = s.productClient.GetProducts(ctx, productIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get products: %w", err)
		}
	}

	productsMap := make(map[int64]string)
	if productsResp != nil {
		for _, p := range productsResp.GetProducts() {
			productsMap[p.GetId()] = p.GetName()
		}
	}

	profile := &dto.UserProfileDTO{
		User: dto.UserResponseDTO{
			ID:    user.GetId(),
			Name:  user.GetName(),
			Email: user.GetEmail(),
		},
		Orders: make([]dto.OrderResponseDTO, 0, len(orders)),
	}

	for _, order := range orders {
		orderDTO := dto.OrderResponseDTO{
			ID: order.GetId(),
			User: dto.UserSummaryDTO{
				ID:    user.GetId(),
				Name:  user.GetName(),
				Email: user.GetEmail(),
			},
			Status:    order.GetStatus(),
			TotalSum:  order.GetTotalAmount(),
			CreatedAt: order.GetCreatedAt().AsTime(),
			Items:     make([]dto.OrderItemDTO, 0, len(order.GetItems())),
		}

		for _, item := range order.GetItems() {
			prodName := "Unknown Product"
			if name, ok := productsMap[item.GetProductId()]; ok {
				prodName = name
			}
			orderDTO.Items = append(orderDTO.Items, dto.OrderItemDTO{
				ProductID:   item.GetProductId(),
				ProductName: prodName,
				Quantity:    item.GetQuantity(),
				UnitPrice:   item.GetPrice(),
			})
		}
		profile.Orders = append(profile.Orders, orderDTO)
	}

	return profile, nil
}
