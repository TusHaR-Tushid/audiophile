package models

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"time"
)

type UserProfile string

const (
	UserRoleAdmin UserProfile = "admin"
	UserRoleUser  UserProfile = "user"
)

type ContextValues struct {
	ID   uuid.UUID `json:"id"`
	Role string    `json:"role"`
}

type UserCredentials struct {
	ID       uuid.UUID `json:"id"`
	Password string    `json:"password"`
	Role     string    `json:"role"`
}

type UsersLoginDetails struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Claims struct {
	ID   uuid.UUID `json:"id"`
	Role string    `json:"role"`
	jwt.StandardClaims
}

type Users struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	PhoneNo  string    `json:"phoneNo"`
	Age      int       `json:"age"`
	UserID   uuid.UUID `json:"userId"`
	Address  string    `json:"address"`
}

type FiltersCheck struct {
	IsSearched   bool
	SearchedName string
	Limit        int
	Page         int
}

type UserDetails struct {
	TotalCount int       `json:"-" db:"total_count"`
	ID         uuid.UUID `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Email      string    `json:"email" db:"email"`
	Password   string    `json:"password" db:"password"`
	PhoneNo    string    `json:"phoneNo" db:"phone_no"`
	Age        int       `json:"age" db:"age"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
	Address    string    `json:"address" db:"address"`
	Role       string    `json:"role"  db:"role"`
}
type TotalUser struct {
	UserDetails []UserDetails
	TotalCount  int `json:"totalCount" db:"total_count"`
}

type Categories struct {
	Name string `json:"name"`
}

type Product struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	Price              float64   `json:"price"`
	Quantity           int       `json:"quantity"`
	BrandID            int       `json:"brandId"`
	ProductDescription string    `json:"productDescription"`
}

type ProductDetails struct {
	TotalCount int       `json:"-" db:"total_count"`
	ID         uuid.UUID `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	CategoryID uuid.UUID `json:"categoryId" db:"category_id"`
	Price      float64   `json:"price" db:"price"`
	Quantity   int       `json:"quantity" db:"quantity"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
	ImageID    uuid.UUID `json:"imageId" db:"image_id"`
	ProductID  uuid.UUID `json:"productId" db:"product_id"`
	URL        string    `json:"url" db:"url"`
}
type ProductUpdateDetails struct {
	Name     string  `json:"name" db:"name"`
	Price    float64 `json:"price" db:"price"`
	Quantity int     `json:"quantity" db:"quantity"`
}
type TotalProduct struct {
	ProductDetails []ProductDetails
	TotalCount     int `json:"totalCount" db:"total_count"`
}

type ProductImages struct {
	ImageID string `json:"imageId"`
}

type CartDetails struct {
	ID uuid.UUID `json:"id"`
}

type OrderDetails struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	AddressID   uuid.UUID `json:"addressId"`
	CartID      []string  `json:"cartId"`
	TotalAmount float64   `json:"totalAmount"`
}

type PaymentDetails struct {
	Name          string `json:"name"`
	PaymentType   string `json:"paymentType"`
	AccountNumber int    `json:"accountNumber"`
}

type BillDetails struct {
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	PaymentID uuid.UUID `json:"paymentId" db:"payment_id"`
	OrderID   uuid.UUID `json:"orderId" db:"order_id"`
}

type Quantity struct {
	NumberOfItems int `json:"numberOfItems"`
}

type Brands struct {
	BrandName        string `json:"brandName"`
	BrandDescription string `json:"brandDescription"`
}
