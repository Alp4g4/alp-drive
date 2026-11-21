package folder

import "errors"

var (
	ErrFolderNameEmpty      = errors.New("folder name can not be empty")
	ErrFolderIDEmpty        = errors.New("folder id can not be empty")
	ErrFolderNotFound       = errors.New("folder not found")
	ErrParentFolderNotFound = errors.New("parent folder not found")
)
