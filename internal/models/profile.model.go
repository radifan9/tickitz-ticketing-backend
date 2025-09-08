package models

// UserProfile represents the user_profiles table
// Note: timestamps are omitted here; add as needed.
type UserProfile struct {
	UserID      string `db:"user_id" json:"user_id"`
	FirstName   string `db:"first_name" json:"first_name,omitempty"`
	LastName    string `db:"last_name" json:"last_name,omitempty"`
	Img         string `db:"img" json:"img,omitempty"`
	PhoneNumber string `db:"phone_number" json:"phone_number,omitempty"`
	Points      int    `db:"points" json:"points,omitempty"`
}
