package models

type Response struct {
	Message string
	Status  string
}

type SuccessResponse struct {
	Success bool `json:"success" example:"true"`
	Status  int  `json:"status" example:"200"`
	Data    any
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Status  int    `json:"status" example:"500"`
	Error   string `json:"error" example:"error message"`
}
