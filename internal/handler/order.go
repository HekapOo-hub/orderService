package handler

import (
	"context"
	"fmt"
	"github.com/HekapOo-hub/orderService/internal/model"
	"github.com/HekapOo-hub/orderService/internal/proto/orderpb"
	"github.com/HekapOo-hub/orderService/internal/proto/positionpb"
	"github.com/HekapOo-hub/orderService/internal/service"
	"github.com/gofrs/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

const (
	PositionPort = ":50005"
	OrderPort    = ":50004"
)

type OrderHandler struct {
	orderpb.UnimplementedOrderServiceServer
	positionClient positionpb.PositionServiceClient
	orderService   service.OrderService
}

func NewOrderHandler(ctx context.Context) (*OrderHandler, error) {
	orderService, err := service.NewOrderService(ctx)
	if err != nil {
		log.Warnf("new order handler: %v", err)
		return nil, fmt.Errorf("new order handler: %w", err)
	}
	conn, err := grpc.Dial(PositionPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Warnf("new order handler: %v", err)
		return nil, fmt.Errorf("new oprder handler: %w", err)
	}
	client := positionpb.NewPositionServiceClient(conn)
	orderHandler := &OrderHandler{orderService: *orderService, positionClient: client}
	go func() {
		for {
			select {
			case <-ctx.Done():
				if err := ctx.Err(); err != nil {
					log.Warnf("position handler get profit and loss: %v", err)
				}
				return
			default:
				err := orderHandler.updatePrices(ctx)
				if err != nil {
					log.Warnf("new order handler: %v", err)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	return orderHandler, nil
}

func (h *OrderHandler) Create(ctx context.Context, order *orderpb.Order) (*orderpb.Empty, error) {
	orderID, err := uuid.NewV4()
	if err != nil {
		log.Warnf("order handler create: %v", err)
		return nil, fmt.Errorf("order handler create: %w", err)
	}
	prices, err := h.orderService.GetPrices()
	if err != nil {
		log.Warnf("order handler create: %v", err)
		return nil, fmt.Errorf("order handler create: %w", err)
	}
	var openPrice float64

	for _, price := range prices {
		if price.Symbol == order.Symbol {
			if order.Side == "BUY" {
				openPrice = price.Ask
			} else {
				openPrice = price.Bid
			}
		}
	}
	err = h.orderService.Create(ctx, model.Order{ID: orderID.String(), Symbol: order.Symbol, AccountID: order.AccountID,
		Price: openPrice, Status: order.Status, Side: order.Side, Time: time.Now().Unix(),
		Leverage: order.Leverage, Quantity: order.Quantity})
	if err != nil {
		log.Warnf("order handler create: %v", err)
		return nil, fmt.Errorf("order handler create: %w", err)
	}
	id, err := uuid.NewV4()
	if err != nil {
		log.Warnf("order handler create: %v", err)
		return nil, fmt.Errorf("order handler create: %w", err)
	}
	_, err = h.positionClient.Open(ctx, &positionpb.Position{ID: id.String(), AccountID: order.AccountID,
		OrderID: orderID.String(), OpenPrice: openPrice, TakeProfit: order.TakeProfit, StopLoss: order.StopLoss,
		GuaranteedStopLoss: order.GuaranteedStopLoss, Quantity: order.Quantity, Symbol: order.Symbol, Leverage: order.Leverage, Side: order.Side})
	if err != nil {
		log.Warnf("order handler create: %v", err)
		return nil, fmt.Errorf("order handler create: %w", err)
	}
	err = h.orderService.SetExecuted(ctx, orderID.String())
	if err != nil {
		log.Warnf("order handler create: %v", err)
		return nil, fmt.Errorf("order handler create: %w", err)
	}
	return &orderpb.Empty{}, nil
}

func (h *OrderHandler) Cancel(ctx context.Context, in *orderpb.OrderID) (*orderpb.Empty, error) {
	err := h.orderService.Cancel(ctx, in.Value)
	if err != nil {
		log.Warnf("order handler cancel: %v", err)
		return nil, fmt.Errorf("order handler cancel: %w", err)
	}
	return &orderpb.Empty{}, nil
}

func (h *OrderHandler) GetOpen(ctx context.Context, in *orderpb.AccountID) (*orderpb.Orders, error) {
	protoOrders := orderpb.Orders{Value: make([]*orderpb.Order, 0)}
	orders, err := h.orderService.GetOpen(ctx, in.Value)
	if err != nil {
		log.Warnf("order handler get open: %v", err)
		return nil, fmt.Errorf("order handler get open: %w", err)
	}
	for _, order := range orders {
		protoOrder := orderpb.Order{ID: order.ID, Symbol: order.Symbol, AccountID: order.AccountID,
			Price: order.Price, Status: order.Status, Side: order.Side, Time: order.Time,
			Leverage: order.Leverage}
		protoOrders.Value = append(protoOrders.Value, &protoOrder)
	}
	return &protoOrders, nil
}

func (h *OrderHandler) updatePrices(ctx context.Context) error {
	prices, err := h.orderService.GetPrices()
	if err != nil {
		return fmt.Errorf("order handler update prices: %w", err)
	}
	stream, err := h.positionClient.UpdatePrices(ctx)
	if err != nil {
		return fmt.Errorf("order handler update prices: %w", err)
	}
	for _, price := range prices {
		err := stream.Send(&positionpb.Price{Symbol: price.Symbol, Ask: price.Ask, Bid: price.Bid})
		if err != nil {
			return fmt.Errorf("order handler update prices: %v", err)
		}
	}
	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("order handler update prices: %v", err)
	}
	return nil
}
