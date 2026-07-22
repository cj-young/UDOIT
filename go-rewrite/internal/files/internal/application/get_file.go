package application

import (
	"context"
	"rewritetest/internal/files/internal/domain"
	"rewritetest/internal/shared/apperr"
)

type GetFileUseCase struct {
	fileRepository domain.FileRepository
}

func NewGetFileUseCase(fileRepo domain.FileRepository) *GetFileUseCase {
	return &GetFileUseCase{
		fileRepository: fileRepo,
	}
}

type GetFileResult struct {
	ID       int64
	CourseID int64
}

func (uc *GetFileUseCase) Execute(ctx context.Context, fileID int64) (GetFileResult, error) {
	file, err := uc.fileRepository.GetFileByID(ctx, fileID)
	if err != nil {
		return GetFileResult{}, err
	}

	if file == nil {
		return GetFileResult{}, apperr.New(
			apperr.CodeNotFound, "file_not_found", "The requested file was not found",
			apperr.WithOp("files.application.get_file.Execute"),
		)
	}

	return GetFileResult{
		ID:       file.ID(),
		CourseID: file.CourseID(),
	}, nil
}
