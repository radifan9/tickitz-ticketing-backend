package models

type User struct {
	Id       string `db:"id" json:"id,omitempty"`
	Role     string `db:"role" json:"role,omitempty"`
	Email    string `db:"email" json:"email,omitempty"`
	Password string `db:"password" json:"password,omitempty"`
}

type RegisterUser struct {
	Email    string `db:"email" json:"email" example:"user@example.com"`
	Password string `db:"password" json:"password" example:"Str0ngP@ss!"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"OldP@ss123"`
	NewPassword string `json:"new_password" binding:"required" example:"NewP@ss456!"`
}


