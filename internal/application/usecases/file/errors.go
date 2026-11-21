package file

import "errors"

var (
	ErrFileNameEmpty     = errors.New("file name can not be empty")
	ErrFileIDEmpty       = errors.New("file id can not be empty")
	ErrFolderIDEmpty     = errors.New("file folder id can not be empty")
	ErrFolderNotFound    = errors.New("folder not found")
	ErrFileContentNil    = errors.New("file content can not be nil")
	ErrFileSizeNegative  = errors.New("file size can not be negative")
	ErrFileMimeTypeEmpty = errors.New("file mime type can not be empty")
)
