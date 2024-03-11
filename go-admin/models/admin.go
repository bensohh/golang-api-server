package models

import "time"

type Teacher struct {
	ID        uint      `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Student struct {
	ID        uint      `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Name      string    `json:"name"`
	Suspended int       `json:"suspended" gorm:"default:0"` // 0: Not suspended or 1: Suspended
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Registry struct {
	ID           uint      `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	TeacherEmail string    `json:"teacher_email" gorm:"primary_key"`
	StudentEmail string    `json:"student_email" gorm:"primary_key"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
