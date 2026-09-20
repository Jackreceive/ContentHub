package models

import "time"

type Article struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int       `gorm:"not null;index" json:"-"`
	Title     string    `gorm:"size:100;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
}
