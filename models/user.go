package models

type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email" gorm:"uniqueIndex"`
	Password  string `json:"password"`
	Name      string `json:"name"`
	UpdatedAt string `json:"-"`
	CreatedAt string `json:"-"`
}
