package folder

import "errors"

var (
	ErrFolderCannotBeOwnParent = errors.New("folder can not be its own parent")
)
