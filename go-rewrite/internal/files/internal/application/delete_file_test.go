package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"rewritetest/internal/files/internal/domain"
	"rewritetest/internal/shared/apperr"
	sharedauth "rewritetest/internal/shared/auth"

	"github.com/stretchr/testify/require"
)

type fakeFileRepository struct {
	file       *domain.File
	getErr     error
	deleteErr  error
	deletedIDs []int64
}

func (f *fakeFileRepository) GetFileByID(context.Context, int64) (*domain.File, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.file, nil
}

func (f *fakeFileRepository) UpdateFile(context.Context, *domain.File) error { return nil }

func (f *fakeFileRepository) DeleteFile(_ context.Context, fileID int64) error {
	f.deletedIDs = append(f.deletedIDs, fileID)
	return f.deleteErr
}

type fakeExternalFileDeleter struct {
	err       error
	called    bool
	fileID    int64
	principal sharedauth.Principal
}

func (f *fakeExternalFileDeleter) DeleteFile(_ context.Context, principal sharedauth.Principal, fileID int64) error {
	f.called = true
	f.fileID = fileID
	f.principal = principal
	return f.err
}

func TestDeleteFileUseCase_FileNotFound(t *testing.T) {
	repo := &fakeFileRepository{}
	deleter := &fakeExternalFileDeleter{}
	uc := NewDeleteFileUseCase(repo, deleter)

	err := uc.Execute(context.Background(), sharedauth.Principal{AgentID: 1, TenantID: 2}, 10)

	require.Error(t, err)
	require.True(t, apperr.IsCode(err, apperr.CodeNotFound))
	require.False(t, deleter.called)
	require.Empty(t, repo.deletedIDs)
}

func TestDeleteFileUseCase_ExternalDeleteErrorBubbles(t *testing.T) {
	repo := &fakeFileRepository{file: domain.NewFile(10, 20, 0, time.Now(), false)}
	deleter := &fakeExternalFileDeleter{err: errors.New("lms unavailable")}
	uc := NewDeleteFileUseCase(repo, deleter)

	err := uc.Execute(context.Background(), sharedauth.Principal{AgentID: 1, TenantID: 2}, 10)

	require.Error(t, err)
	require.EqualError(t, err, "lms unavailable")
	require.True(t, deleter.called)
	require.Empty(t, repo.deletedIDs)
}

func TestDeleteFileUseCase_LocalDeleteErrorBubbles(t *testing.T) {
	repo := &fakeFileRepository{file: domain.NewFile(10, 20, 0, time.Now(), false), deleteErr: errors.New("db delete failed")}
	deleter := &fakeExternalFileDeleter{}
	uc := NewDeleteFileUseCase(repo, deleter)

	err := uc.Execute(context.Background(), sharedauth.Principal{AgentID: 1, TenantID: 2}, 10)

	require.Error(t, err)
	require.EqualError(t, err, "db delete failed")
	require.True(t, deleter.called)
	require.Equal(t, []int64{10}, repo.deletedIDs)
}

func TestDeleteFileUseCase_Success(t *testing.T) {
	repo := &fakeFileRepository{file: domain.NewFile(10, 20, 0, time.Now(), false)}
	deleter := &fakeExternalFileDeleter{}
	uc := NewDeleteFileUseCase(repo, deleter)
	principal := sharedauth.Principal{AgentID: 99, TenantID: 3}

	err := uc.Execute(context.Background(), principal, 10)

	require.NoError(t, err)
	require.True(t, deleter.called)
	require.Equal(t, int64(10), deleter.fileID)
	require.Equal(t, principal, deleter.principal)
	require.Equal(t, []int64{10}, repo.deletedIDs)
}
