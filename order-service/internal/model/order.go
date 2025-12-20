package model

import "time"

type Order struct {
	ID          int64       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64       `gorm:"index;not null" json:"user_id"`
	Status      string      `gorm:"type:varchar(50);default:'created';not null" json:"status"`
	TotalAmount float64     `gorm:"type:decimal(15,2);not null" json:"total_amount"`
	Items       []OrderItem `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"items"`
	CreatedAt   time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Order) TableName() string {
	return "orders"
}
