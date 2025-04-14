package models

import "errors"

var (
	ErrNoActiveReception     = errors.New("no active reception")
	ErrInvalidProductType    = errors.New("invalid product type")
	ErrInvalidRequest        = errors.New("invalid request")
	ErrInvalidRole           = errors.New("invalid role")
	ErrTokenGenerationFailed = errors.New("failed to generate token")
	ErrInvalidCity           = errors.New("invalid city")
)
