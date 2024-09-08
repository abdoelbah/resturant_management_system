package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"password" db:"password"`
	Img       string    `json:"img,omitempty" db:"img"`
	Phone     string    `json:"phone,omitempty" db:"phone"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Role struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type UserRole struct {
	UserID uuid.UUID `json:"user_id" db:"user_id"`
	RoleID int       `json:"role_id" db:"role_id"`
}

type Item struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Img       string    `json:"img,omitempty" db:"img"`
	Price     float64   `json:"price" db:"price"`
	VendorID  uuid.UUID `json:"vendor_id" db:"vendor_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Cart struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	TotalPrice float64   `json:"total_price" db:"total_price"`
	Quantity   int       `json:"quantity" db:"quantity"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type CartItem struct {
	CartID   uuid.UUID `json:"cart_id" db:"cart_id"`
	ItemID   uuid.UUID `json:"item_id" db:"item_id"`
	Quantity int       `json:"quantity" db:"quantity"`
}

type Order struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrderTotalCost float64   `json:"order_total_cost" db:"order_total_cost"`
	CartID         uuid.UUID `json:"cart_id" db:"cart_id"`
	CustomerID     uuid.UUID `json:"customer_id" db:"customer_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
	OrderID  uuid.UUID `json:"order_id" db:"order_id"`
	ItemID   uuid.UUID `json:"item_id" db:"item_id"`
	Quantity int       `json:"quantity" db:"quantity"`
	Price    float64   `json:"price" db:"price"`
}

type Vendor struct {
	ID          uuid.UUID `json:"ID" db:"id"`
	Name        string    `json:"Name" db:"name"`
	Email       string    `json:"Email" db:"email"`
	Phone       string    `json:"Phone" db:"phone"`
	Img         string    `json:"Img" db:"img"`
	Description string    `json:"Description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
