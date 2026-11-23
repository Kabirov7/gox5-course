package domain

import "errors"

var ErrNotFound = errors.New("user not found")

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type CreateUser struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}
