package model

type OrderItem struct {
	ID          int64   `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID     int64   `gorm:"index;not null" json:"order_id"`
	ProductID   int64   `gorm:"not null" json:"product_id"`
	ProductName string  `gorm:"not null" json:"product_name"`
	Price       float64 `gorm:"type:decimal(15,2);not null" json:"price"`
	Quantity    int32   `gorm:"not null" json:"quantity"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
