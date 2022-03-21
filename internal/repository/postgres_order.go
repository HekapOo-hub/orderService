package repository

import (
	"context"
	"fmt"
	"github.com/HekapOo-hub/orderService/internal/config"
	"github.com/HekapOo-hub/orderService/internal/model"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"
)

type OrderRepository interface {
	Create(ctx context.Context, order model.Order) error
	Delete(ctx context.Context, orderID string) error
	GetOpen(ctx context.Context, accountID string) ([]model.Order, error)
	SetExecuted(ctx context.Context, orderID string) error
}
type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(ctx context.Context) (*PostgresOrderRepository, error) {
	cfg, err := config.GetPostgresConfig()
	if err != nil {
		return nil, fmt.Errorf("new postgres repository func %v", err)
	}

	pool, err := pgxpool.Connect(context.Background(), cfg.GetURL())
	if err != nil {
		return nil, fmt.Errorf("creating postgres connection pool error %w", err)
	}
	orderRepository := &PostgresOrderRepository{pool: pool}
	return orderRepository, nil
}

func (repo *PostgresOrderRepository) Create(ctx context.Context, order model.Order) error {
	query := "insert into orders (symbol,id,accountID,price,status,side,time,leverage,quantity) values ($1,$2,$3,$4,$5,$6,$7,$8,$9)"
	_, err := repo.pool.Exec(ctx, query, order.Symbol, order.ID, order.AccountID, order.Price, "NEW", order.Side, order.Time,
		order.Leverage, order.Quantity)
	if err != nil {
		log.Infof("%v", order)
		return fmt.Errorf("postgres create order: %w", err)
	}

	return nil
}

func (repo *PostgresOrderRepository) Delete(ctx context.Context, orderID string) error {
	rowsAffected, err := repo.pool.Exec(ctx, "delete from orders where status=$1 && id=$2", "NEW", orderID)
	if err != nil {
		return fmt.Errorf("postgres delete order %w", err)
	}
	if rowsAffected.RowsAffected() == 0 {
		return fmt.Errorf("no such open order in db")
	}
	return nil
}

func (repo *PostgresOrderRepository) GetOpen(ctx context.Context, accountID string) ([]model.Order, error) {
	query := "select * from orders where status=$1 && accountID=$2"
	rows, err := repo.pool.Query(ctx, query, "NEW", accountID)
	if err != nil {
		return nil, fmt.Errorf("posgres get open orders: %w", err)
	}
	defer rows.Close()
	openOrders := make([]model.Order, 0)
	for rows.Next() {
		var o model.Order
		err = rows.Scan(&o.ID, &o.Symbol, &o.AccountID, &o.Price, &o.Status, &o.Side, &o.Time, &o.Leverage, &o.Quantity)
		if err != nil {
			return nil, fmt.Errorf("postgres get open orders: %w", err)
		}
		openOrders = append(openOrders, o)
	}
	return openOrders, nil
}

func (repo *PostgresOrderRepository) SetExecuted(ctx context.Context, orderID string) error {
	query := "update orders set status=$1 where id=$2"
	_, err := repo.pool.Exec(ctx, query, "EXECUTED", orderID)
	if err != nil {
		return fmt.Errorf("postgres update order %w", err)
	}
	return nil
}
