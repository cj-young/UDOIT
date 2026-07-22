package application

import (
	"context"
	"rewritetest/internal/files/internal/domain"
	"rewritetest/internal/shared/apperr"
	"rewritetest/internal/shared/auth"
)

type DeleteFileUseCase struct {
	fileRepository domain.FileRepository
	fileDeleter    FileDeleter
}

type FileDeleter interface {
	DeleteFile(ctx context.Context, principal auth.Principal, fileID int64) error
}

func NewDeleteFileUseCase(fileRepository domain.FileRepository, fileDeleter FileDeleter) *DeleteFileUseCase {
	return &DeleteFileUseCase{
		fileRepository: fileRepository,
		fileDeleter:    fileDeleter,
	}
}

func (u *DeleteFileUseCase) Execute(ctx context.Context, principal auth.Principal, fileID int64) error {

	file, err := u.fileRepository.GetFileByID(ctx, fileID)
	if err != nil {
		return err
	}

	if file == nil {
		return apperr.New(
			apperr.CodeNotFound, "file_not_found", "The file could not be deleted because no file was found.",
			apperr.WithOp("files.application.delete_file.Execute"),
		)
	}

	err = u.fileDeleter.DeleteFile(ctx, principal, fileID)
	if err != nil {
		return err
	}

	return u.fileRepository.DeleteFile(ctx, file.ID())
}
