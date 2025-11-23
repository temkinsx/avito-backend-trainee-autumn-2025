package dto

type ErrorResponseDTO struct {
	Error ErrorDTO `json:"error"`
}

type ErrorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
