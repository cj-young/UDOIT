package scanner

import (
	"database/sql"
	"net/http"
	"rewritetest/internal/scanner/internal"
	"rewritetest/internal/scanner/internal/application"
	"rewritetest/internal/scanner/internal/infrastructure"

	"github.com/gin-gonic/gin"
)


type Module struct {
	handler *internal.Handler
	authenticator internal.Authenticator
}

func New(
	db *sql.DB,
	courseRetriever application.CourseRetriever,
	contentItemService application.ContentItemService,
	externalContentReceiver application.ExternalContentRetriever,
	issueService application.IssueService,
	issueRetriever application.IssueRetriever,
	authenticator internal.Authenticator,
) *Module {

	httpClient := &http.Client{}
	baseURL := "http://udoit3-ace:3000"
	scanner := infrastructure.NewEqualAccessScanner(httpClient, baseURL)

	blake3Hasher := infrastructure.NewBlake3ContentHasher()
	reportRepository := infrastructure.NewMySQLReportRepository(db)

	scanCourseUseCase := application.NewScanCourseUseCase(courseRetriever, contentItemService, externalContentReceiver, blake3Hasher, issueService, scanner)
	createReportUseCase := application.NewCreateReportUseCase(reportRepository, issueRetriever)
	handler := internal.NewHandler(scanCourseUseCase, createReportUseCase)

	return &Module{
		handler: handler,
		authenticator: authenticator,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	m.handler.RegisterRoutes(rg, m.authenticator)
}