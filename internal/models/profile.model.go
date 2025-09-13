package models

import (
	"mime/multipart"
	"time"
)

// UserProfile represents the user_profiles table
type UserProfile struct {
	UserID      string    `db:"user_id" json:"user_id,omitempty"`
	FirstName   string    `db:"first_name" json:"first_name,omitempty" form:"first_name"`
	LastName    string    `db:"last_name" json:"last_name,omitempty"`
	Img         string    `db:"img" json:"img,omitempty"`
	PhoneNumber string    `db:"phone_number" json:"phone_number,omitempty"`
	Points      int       `db:"points" json:"points,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

// type EditUserProfile struct {
// 	UserProfile
// 	Img *multipart.FileHeader `form:"img"`
// }

type EditUserProfile struct {
	FirstName   string                `form:"first_name"`
	LastName    string                `form:"last_name"`
	PhoneNumber string                `form:"phone_number"`
	Img         *multipart.FileHeader `form:"img"`
}
