package models

import (
	"github.com/google/uuid"
)

type Photo struct {
	ID      uuid.UUID
	Bytes   string // base64
	Preview string // base64
}
