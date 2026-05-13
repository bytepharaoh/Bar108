package repository

import (
	"bar108/internal/apperror"
	"bar108/internal/db"
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// =============================================
// Input / Output structs
// =============================================

// OrderItemInput is one item line from the customer's request.
// Customer sends menu_item_id + quantity only.
// We look up the price ourselves — customer never sends price.
// This prevents price manipulation attacks.
type OrderItemInput struct {
	MenuItemID int32
	Quantity   int32
}

// PlaceOrderInput groups everything needed to place a full order.
type PlaceOrderInput struct {
	UserID          int32
	Items           []OrderItemInput
	PromoCode       string
	DeliveryAddress string
	Notes           string
}

// PlaceOrderResult is everything the handler needs to build the response.
type PlaceOrderResult struct {
	Order      db.Order
	Items      []db.OrderItem
	TotalPrice string
	FinalPrice string
}

// =============================================
// Interface — consumed by the order service.
// The service will define its own private orderStore
// interface with only the methods it needs.
// =============================================

type OrderRepository interface {
	PlaceOrder(ctx context.Context, input PlaceOrderInput) (PlaceOrderResult, error)
	GetOrderByID(ctx context.Context, id int32) (db.GetOrderByIDRow, error)
	GetOrdersByUserID(ctx context.Context, userID int32) ([]db.GetOrdersByUserIDRow, error)
	GetAllOrders(ctx context.Context) ([]db.GetAllOrdersRow, error)
	GetPendingOrders(ctx context.Context) ([]db.GetPendingOrdersRow, error)
	GetOrderItems(ctx context.Context, orderID int32) ([]db.GetOrderItemsRow, error)
	GetOrderStatusHistory(ctx context.Context, orderID int32) ([]db.OrderStatusHistory, error)
	UpdateOrderStatus(ctx context.Context, id int32, status string) (db.Order, error)
	AssignCourier(ctx context.Context, orderID int32, courierID int32) (db.Order, error)
	CancelOrder(ctx context.Context, id int32) (db.Order, error)
	GetAllCouriers(ctx context.Context) ([]db.Courier, error)
	GetAvailableCouriers(ctx context.Context) ([]db.Courier, error)
	GetCourierByID(ctx context.Context, id int32) (db.Courier, error)
	UpdateCourierStatus(ctx context.Context, id int32, status string) (db.Courier, error)
}

// =============================================
// Concrete struct — unexported, internal detail.
// NewOrderRepository returns the interface so
// external packages never reference this type by name.
// =============================================

type orderRepository struct {
	db      *sql.DB     // raw DB — needed to call BeginTx()
	queries *db.Queries // SQLC queries — for non-transactional ops
}

// NewOrderRepository returns the interface, not the concrete type.
// This means wire.go can use OrderRepository as the type —
// no unexported type leaking outside the package.
func NewOrderRepository(conn *sql.DB) OrderRepository {
	return &orderRepository{
		db:      conn,
		queries: db.New(conn),
	}
}

// =============================================
// PlaceOrder — the transaction
// All 9 steps succeed together or all fail together.
// =============================================

func (r *orderRepository) PlaceOrder(ctx context.Context, input PlaceOrderInput) (PlaceOrderResult, error) {
	// Start the transaction.
	// BeginTx creates a private workspace in PostgreSQL.
	// Changes here are invisible to everyone else until Commit().
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return PlaceOrderResult{}, fmt.Errorf("PlaceOrder begin tx: %w", err)
	}

	// Safety net — if we return early due to any error,
	// Rollback fires automatically and erases all partial changes.
	// After Commit() succeeds, Rollback() becomes a no-op.
	defer func() {
		_ = tx.Rollback()
	}()

	// Bind all SQLC queries to this transaction connection.
	// Every query through qtx participates in the same atomic unit.
	// Using r.queries here instead would run queries OUTSIDE the transaction.
	qtx := r.queries.WithTx(tx)

	// STEP 1 — Validate items + calculate total price.
	// Inside the transaction so PostgreSQL locks the rows we read —
	// no item can be deleted between our check and our INSERT.
	var totalPrice float64
	var itemParams []db.CreateOrderItemParams

	for _, item := range input.Items {
		menuItem, err := qtx.GetMenuItemByID(ctx, item.MenuItemID)
		if err != nil {
			if err == sql.ErrNoRows {
				return PlaceOrderResult{}, apperror.ErrMenuNotFound
			}
			return PlaceOrderResult{}, fmt.Errorf("PlaceOrder get menu item: %w", err)
		}

		// Business rule: unavailable items cannot be ordered.
		if !menuItem.Available {
			return PlaceOrderResult{}, apperror.ErrItemUnavailable
		}

		// strconv.ParseFloat — correct tool for string → float64.
		// Returns an error we can handle, unlike fmt.Sscanf.
		// 64 = we want float64 precision.
		price, err := strconv.ParseFloat(menuItem.Price, 64)
		if err != nil {
			return PlaceOrderResult{}, fmt.Errorf("PlaceOrder parse price for item %d: %w", item.MenuItemID, err)
		}

		totalPrice += price * float64(item.Quantity)

		itemParams = append(itemParams, db.CreateOrderItemParams{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
			// Price snapshot — locks in the price at ordering time.
			// If menu price changes tomorrow, this order still shows
			// what the customer actually paid.
			UnitPrice: menuItem.Price,
		})
	}

	// STEP 2 — Apply promo code if provided.
	var discountAmount float64
	var promotionID sql.NullInt32

	if input.PromoCode != "" {
		promo, err := qtx.GetPromotionByCode(ctx, input.PromoCode)
		if err != nil {
			if err == sql.ErrNoRows {
				return PlaceOrderResult{}, apperror.ErrPromoNotFound
			}
			return PlaceOrderResult{}, fmt.Errorf("PlaceOrder get promo: %w", err)
		}

		// Check expiry — ExpiresAt.Valid means field is NOT NULL.
		// Some promos have no expiry — unlimited time.
		// time.Now() is the correct way — never parse from context.
		if promo.ExpiresAt.Valid && promo.ExpiresAt.Time.Before(time.Now()) {
			return PlaceOrderResult{}, apperror.ErrPromoExpired
		}

		// Check usage limit — UsageLimit.Valid means field is NOT NULL.
		// Some promos have no usage cap — unlimited uses.
		if promo.UsageLimit.Valid && promo.UsedCount >= promo.UsageLimit.Int32 {
			return PlaceOrderResult{}, apperror.ErrPromoExhausted
		}

		// Parse discount value — strconv, not fmt.Sscanf.
		promoValue, err := strconv.ParseFloat(promo.DiscountValue, 64)
		if err != nil {
			return PlaceOrderResult{}, fmt.Errorf("PlaceOrder parse promo value: %w", err)
		}

		switch promo.DiscountType {
		case "percentage":
			// e.g. 10% off 350.00 → 35.00 discount
			discountAmount = totalPrice * (promoValue / 100)
		case "fixed":
			// e.g. 50 RUB off — but never more than the total
			discountAmount = promoValue
			if discountAmount > totalPrice {
				discountAmount = totalPrice
			}
		}

		promotionID = sql.NullInt32{Int32: promo.ID, Valid: true}

		// Increment usage count INSIDE the transaction.
		// If anything fails later, this rolls back too —
		// the promo code usage count stays accurate.
		if _, err := qtx.IncrementPromotionUsage(ctx, promo.ID); err != nil {
			return PlaceOrderResult{}, fmt.Errorf("PlaceOrder increment promo usage: %w", err)
		}
	}

	// STEP 3 — Calculate final price.
	finalPrice := totalPrice - discountAmount

	// strconv.FormatFloat — converts float64 back to string.
	// 'f' = decimal notation (not scientific like 3.5e+2).
	//  2  = 2 decimal places.
	// 64  = float64 source precision.
	totalStr := strconv.FormatFloat(totalPrice, 'f', 2, 64)
	discountStr := strconv.FormatFloat(discountAmount, 'f', 2, 64)
	finalStr := strconv.FormatFloat(finalPrice, 'f', 2, 64)

	// STEP 4 — Insert the order row.
	order, err := qtx.CreateOrder(ctx, db.CreateOrderParams{
		UserID:         input.UserID,
		PromotionID:    promotionID,
		Status:         "pending",
		TotalPrice:     totalStr,
		DiscountAmount: discountStr,
		FinalPrice:     finalStr,
		Notes: sql.NullString{
			String: input.Notes,
			Valid:  input.Notes != "",
		},
		DeliveryAddress: sql.NullString{
			String: input.DeliveryAddress,
			Valid:  input.DeliveryAddress != "",
		},
	})
	if err != nil {
		return PlaceOrderResult{}, fmt.Errorf("PlaceOrder create order: %w", err)
	}

	// STEP 5 — Insert all order item rows.
	// Now we have the real order.ID to attach them to.
	var createdItems []db.OrderItem
	for _, params := range itemParams {
		params.OrderID = order.ID
		item, err := qtx.CreateOrderItem(ctx, params)
		if err != nil {
			return PlaceOrderResult{}, fmt.Errorf("PlaceOrder create order item: %w", err)
		}
		createdItems = append(createdItems, item)
	}

	// STEP 6 — Record first status history entry.
	// This is the start of the timeline the customer will track.
	// Every future status change adds another row here.
	_, err = qtx.CreateOrderStatusHistory(ctx, db.CreateOrderStatusHistoryParams{
		OrderID: order.ID,
		Status:  "pending",
		Note:    sql.NullString{String: "Order placed successfully", Valid: true},
	})
	if err != nil {
		return PlaceOrderResult{}, fmt.Errorf("PlaceOrder create status history: %w", err)
	}

	// STEP 7 — Commit.
	// Only reached if ALL steps above succeeded.
	// Makes all changes permanent and visible to the rest of the system.
	if err := tx.Commit(); err != nil {
		return PlaceOrderResult{}, fmt.Errorf("PlaceOrder commit: %w", err)
	}

	return PlaceOrderResult{
		Order:      order,
		Items:      createdItems,
		TotalPrice: totalStr,
		FinalPrice: finalStr,
	}, nil
}

// =============================================
// Single query operations
// No transaction needed — each is one atomic SQL statement.
// =============================================

func (r *orderRepository) GetOrderByID(ctx context.Context, id int32) (db.GetOrderByIDRow, error) {
	order, err := r.queries.GetOrderByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.GetOrderByIDRow{}, apperror.ErrOrderNotFound
		}
		return db.GetOrderByIDRow{}, fmt.Errorf("GetOrderByID: %w", err)
	}
	return order, nil
}

func (r *orderRepository) GetOrdersByUserID(ctx context.Context, userID int32) ([]db.GetOrdersByUserIDRow, error) {
	orders, err := r.queries.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetOrdersByUserID: %w", err)
	}
	return orders, nil
}

func (r *orderRepository) GetAllOrders(ctx context.Context) ([]db.GetAllOrdersRow, error) {
	orders, err := r.queries.GetAllOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllOrders: %w", err)
	}
	return orders, nil
}

func (r *orderRepository) GetPendingOrders(ctx context.Context) ([]db.GetPendingOrdersRow, error) {
	orders, err := r.queries.GetPendingOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetPendingOrders: %w", err)
	}
	return orders, nil
}

func (r *orderRepository) GetOrderItems(ctx context.Context, orderID int32) ([]db.GetOrderItemsRow, error) {
	items, err := r.queries.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("GetOrderItems: %w", err)
	}
	return items, nil
}

func (r *orderRepository) GetOrderStatusHistory(ctx context.Context, orderID int32) ([]db.OrderStatusHistory, error) {
	history, err := r.queries.GetOrderStatusHistory(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("GetOrderStatusHistory: %w", err)
	}
	return history, nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, id int32, status string) (db.Order, error) {
	order, err := r.queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return db.Order{}, apperror.ErrOrderNotFound
		}
		return db.Order{}, fmt.Errorf("UpdateOrderStatus: %w", err)
	}
	return order, nil
}

func (r *orderRepository) AssignCourier(ctx context.Context, orderID int32, courierID int32) (db.Order, error) {
	order, err := r.queries.AssignCourier(ctx, db.AssignCourierParams{
		ID:        orderID,
		CourierID: sql.NullInt32{Int32: courierID, Valid: true},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return db.Order{}, apperror.ErrOrderNotFound
		}
		return db.Order{}, fmt.Errorf("AssignCourier: %w", err)
	}
	return order, nil
}

func (r *orderRepository) CancelOrder(ctx context.Context, id int32) (db.Order, error) {
	order, err := r.queries.CancelOrder(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.Order{}, apperror.ErrOrderNotFound
		}
		return db.Order{}, fmt.Errorf("CancelOrder: %w", err)
	}
	return order, nil
}

func (r *orderRepository) GetAllCouriers(ctx context.Context) ([]db.Courier, error) {
	couriers, err := r.queries.GetAllCouriers(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllCouriers: %w", err)
	}
	return couriers, nil
}

func (r *orderRepository) GetAvailableCouriers(ctx context.Context) ([]db.Courier, error) {
	couriers, err := r.queries.GetAvailableCouriers(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAvailableCouriers: %w", err)
	}
	return couriers, nil
}

func (r *orderRepository) GetCourierByID(ctx context.Context, id int32) (db.Courier, error) {
	courier, err := r.queries.GetCourierByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.Courier{}, apperror.ErrNotFound
		}
		return db.Courier{}, fmt.Errorf("GetCourierByID: %w", err)
	}
	return courier, nil
}

func (r *orderRepository) UpdateCourierStatus(ctx context.Context, id int32, status string) (db.Courier, error) {
	courier, err := r.queries.UpdateCourierStatus(ctx, db.UpdateCourierStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return db.Courier{}, apperror.ErrNotFound
		}
		return db.Courier{}, fmt.Errorf("UpdateCourierStatus: %w", err)
	}
	return courier, nil
}
