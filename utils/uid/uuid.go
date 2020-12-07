package uid

import (
	"encoding/base64"

	"github.com/satori/go.uuid"
)

func NewId() string {
	id := uuid.NewV4()
	b64 := base64.URLEncoding.EncodeToString(id.Bytes()[:12])
	return b64
}
