package service

import (
	"context"
	"fmt"

	"github.com/microserviceteam0/bff-gateway/bff/internal/dto"
	orderv1 "github.com/microserviceteam0/bff-gateway/bff/api/proto/order/v1"
	productv1 "github.com/microserviceteam0/bff-gateway/bff/api/proto/product"
	userv1 "github.com/microserviceteam0/bff-gateway/bff/api/proto/user"
	"golang.org/x/sync/errgroup"
)

func (s *bffService) CreateOrder(ctx context.Context, userID int64, userRole string, req dto.CreateOrderRequestDTO) (*dto.OrderResponseDTO, error) {
	items := make([]*orderv1.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = &orderv1.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	createReq := &orderv1.CreateOrderRequest{
		UserId: userID,
		Items:  items,
	}

	ctx = withAuthMetadata(ctx, userID, userRole)
	resp, err := s.orderClient.CreateOrder(ctx, createReq)
	if err != nil {
		return nil, err
	}

	return &dto.OrderResponseDTO{
		ID:     resp.GetOrderId(),
		Status: "pending",
	}, nil
}

func (s *bffService) CancelOrder(ctx context.Context, userID int64, userRole string, orderID int64, reason string) error {
	ctx = withAuthMetadata(ctx, userID, userRole)
	_, err := s.orderClient.CancelOrder(ctx, orderID, userID, reason)
	return err
}

func (s *bffService) GetOrderDetails(ctx context.Context, userID int64, userRole string, orderID int64) (*dto.OrderResponseDTO, error) {
	ctx = withAuthMetadata(ctx, userID, userRole)
	
	orderResp, err := s.orderClient.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	order := orderResp.GetOrder()

	var (
		userResp     *userv1.GetUserResponse
		productsResp *productv1.ProductsResponse
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		userResp, err = s.userClient.GetUser(ctx, order.GetUserId())
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		productIDs := make([]int64, 0, len(order.GetItems()))
		for _, item := range order.GetItems() {
			productIDs = append(productIDs, item.GetProductId())
		}

		if len(productIDs) == 0 {
			productsResp = &productv1.ProductsResponse{Products: []*productv1.ProductResponse{}}
			return nil
		}

		var err error
		productsResp, err = s.productClient.GetProducts(ctx, productIDs)
		if err != nil {
			return fmt.Errorf("failed to get products: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	user := userResp.GetUser()
	
	productsMap := make(map[int64]string)
	if productsResp != nil {
		for _, p := range productsResp.GetProducts() {
			productsMap[p.GetId()] = p.GetName()
		}
	}

	resp := &dto.OrderResponseDTO{
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

		resp.Items = append(resp.Items, dto.OrderItemDTO{
			ProductID:   item.GetProductId(),
			ProductName: prodName,
			Quantity:    item.GetQuantity(),
			UnitPrice:   item.GetPrice(),
		})
	}

	return resp, nil
}
