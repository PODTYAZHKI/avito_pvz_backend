package models

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID
	Email    string
	Password string
	Role     string
}

type TokenResponse struct {
	Token string `json:"token"`
}

var (
	RoleModerator = "moderator"
	RoleEmployee  = "employee"
)
