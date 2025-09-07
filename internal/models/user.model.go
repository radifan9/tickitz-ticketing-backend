package models

type User struct {
	Id       string `db:"id" json:"id,omitempty"`
	Role     string `db:"role" json:"role,omitempty"`
	Email    string `db:"email" json:"email,omitempty"`
	Password string `db:"password" json:"password,omitempty"`
}

type RegisterUser struct {
	Email    string `db:"email" json:"email"`
	Password string `db:"password" json:"password"`
}
