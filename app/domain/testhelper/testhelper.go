package testhelper

import "github.com/google/uuid"

func RandomString() string {
	return uuid.New().String()
}
