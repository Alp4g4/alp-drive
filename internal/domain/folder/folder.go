package folder

import (
	"time"
	"uuid"
)

type Folder struct {
	ID        uuid.UUID
	Name      string
	Parent    *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (f *Folder) Move(parentID *uuid.UUID) error {
	if parentID != nil && *parentID == f.ID {
		return ErrFolderCannotBeOwnParent
	}

	f.Parent = parentID

	return nil
}
