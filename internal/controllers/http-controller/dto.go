package httpcontroller

import (
	"github.com/google/uuid"
)

type photoDTO struct {
	ID      uuid.UUID `json:"id"`
	Bytes   string    `json:"bytes"`   // base64
	Preview string    `json:"preview"` // base64
}
