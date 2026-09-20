package models

import "time"

type User struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"size:50;not null;uniqueIndex" json:"email"`
	Password  string    `gorm:"szie:255;not null" json:"password"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	UpdatedAt time.Time `json:"-"`
	CreatedAt time.Time `json:"-"`
}
