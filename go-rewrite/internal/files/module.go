package files

import (
	"database/sql"
	"rewritetest/internal/files/internal"
	"rewritetest/internal/files/internal/application"
	"rewritetest/internal/files/internal/infrastructure"

	"github.com/gin-gonic/gin"
)

type Module struct {
	handler *internal.Handler
}

func New(db *sql.DB, fileDeleter application.LMSFileDeleter) *Module {

	fileRepository := infrastructure.NewMySQLFileRepository(db)
	getFileUseCase := application.NewGetFileUseCase(fileRepository)
	deleteFileUseCase := application.NewDeleteFileUseCase(fileRepository, fileDeleter)
	handler := internal.NewHandler(getFileUseCase, deleteFileUseCase)

	return &Module{
		handler: handler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	m.handler.RegisterRoutes(rg)
}