package domain

import "context"

type FileRepository interface {
	GetFileByID(ctx context.Context, fileID int64) (*File, error)
	UpdateFile(ctx context.Context, file *File) error
	DeleteFile(ctx context.Context, fileID int64) error
}
