package file

import (
	"time"
	"uuid"
)

type File struct {
	ID        uuid.UUID
	Name      string
	Folder    uuid.UUID
	Size      int64
	MimeType  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
