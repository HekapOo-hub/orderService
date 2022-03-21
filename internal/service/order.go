package service

import (
	"context"
	"fmt"
	"github.com/HekapOo-hub/orderService/internal/model"
	"github.com/HekapOo-hub/orderService/internal/repository"
)

type OrderService struct {
	priceUpdates    repository.PriceUpdates
	orderRepository repository.OrderRepository
}

func NewOrderService(ctx context.Context) (*OrderService, error) {
	priceUpdates, err := repository.NewRedisPriceUpdates(ctx)
	if err != nil {
		return nil, fmt.Errorf("new order service: %w", err)
	}
	repo, err := repository.NewOrderRepository(ctx)
	if err != nil {
		return nil, fmt.Errorf("new order sercvice: %w", err)
	}
	return &OrderService{priceUpdates: priceUpdates, orderRepository: repo}, nil
}

func (s *OrderService) Create(ctx context.Context, order model.Order) error {
	err := s.orderRepository.Create(ctx, order)
	if err != nil {
		return fmt.Errorf("order service create: %w", err)
	}
	return nil
}

func (s *OrderService) Cancel(ctx context.Context, orderID string) error {
	err := s.orderRepository.Delete(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order service cancel: %w", err)
	}
	return nil
}

func (s *OrderService) GetOpen(ctx context.Context, accountID string) ([]model.Order, error) {
	openOrders, err := s.orderRepository.GetOpen(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("order service get open: %w", err)
	}
	return openOrders, nil
}

func (s *OrderService) SetExecuted(ctx context.Context, orderID string) error {
	err := s.orderRepository.SetExecuted(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order service set executed: %w", err)
	}
	return nil
}

func (s *OrderService) GetPrices() ([]model.GeneratedPrice, error) {
	pricesMap, err := s.priceUpdates.Get()
	if err != nil {
		return nil, fmt.Errorf("order srervice get: %w", err)
	}
	prices := make([]model.GeneratedPrice, 0)
	for _, price := range pricesMap {
		prices = append(prices, price)
	}
	return prices, nil
}
