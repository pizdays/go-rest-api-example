package util

import (
	uuid "github.com/satori/go.uuid"
)

func GenerateUUIDv4() string {
	id, _ := uuid.NewV4()
	return id.String()
}
