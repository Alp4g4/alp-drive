package folder

import (
	"errors"
	"testing"
	"uuid"
)

func TestFolder_Move(t *testing.T) {
	id := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b80")
	otherParentID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b81")

	t.Run("cannot be its own parent", func(t *testing.T) {
		f := Folder{ID: id, Parent: &otherParentID}

		err := f.Move(&id)
		if !errors.Is(err, ErrFolderCannotBeOwnParent) {
			t.Errorf("err = %v, want ErrFolderCannotBeOwnParent", err)
		}
		if f.Parent != &otherParentID {
			t.Errorf("Parent was changed despite the error")
		}
	})

	t.Run("moved under a new parent", func(t *testing.T) {
		f := Folder{ID: id}

		if err := f.Move(&otherParentID); err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if f.Parent == nil || *f.Parent != otherParentID {
			t.Errorf("Parent = %v, want %v", f.Parent, otherParentID)
		}
	})

	t.Run("moved to root", func(t *testing.T) {
		f := Folder{ID: id, Parent: &otherParentID}

		if err := f.Move(nil); err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if f.Parent != nil {
			t.Errorf("Parent = %v, want nil", f.Parent)
		}
	})
}
