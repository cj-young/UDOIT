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

func New(db *sql.DB) *Module {

	fileRepository := infrastructure.NewMySQLFileRepository(db)
	getFileUseCase := application.NewGetFileUseCase(fileRepository)
	handler := internal.NewHandler(getFileUseCase)

	return &Module{
		handler: handler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	m.handler.RegisterRoutes(rg)
}