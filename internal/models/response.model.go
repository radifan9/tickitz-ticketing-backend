package models

type SuccessResponse struct {
	Success bool `json:"success" example:"true"`
	Status  int  `json:"status" example:"200"`
	Data    any  `json:"data"`
}

type SuccessLoginResponse struct {
	Role  string `json:"role"`
	Token string `json:"token"`
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Status  int    `json:"status" example:"500"`
	Error   string `json:"error" example:"error message"`
}
