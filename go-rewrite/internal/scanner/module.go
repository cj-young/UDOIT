package scanner

import (
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
	courseRetriever application.CourseRetriever,
	contentItemService application.ContentItemService,
	externalContentReceiver application.ExternalContentRetriever,
	issueService application.IssueService,
	authenticator internal.Authenticator,
) *Module {

	httpClient := &http.Client{}
	baseURL := "http://udoit3-ace:3000"
	scanner := infrastructure.NewEqualAccessScanner(httpClient, baseURL)

	blake3Hasher := infrastructure.NewBlake3ContentHasher()

	scanCourseUseCase := application.NewScanCourseUseCase(courseRetriever, contentItemService, externalContentReceiver, blake3Hasher, issueService, scanner)
	handler := internal.NewHandler(scanCourseUseCase)

	return &Module{
		handler: handler,
		authenticator: authenticator,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	m.handler.RegisterRoutes(rg, m.authenticator)
}