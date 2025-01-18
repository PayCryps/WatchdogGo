package db

type User struct {
	ID    string `gorm:"primary_key"`
	Name  string `gorm:"not null"`
	Email string `gorm:"unique;not null"`
}
