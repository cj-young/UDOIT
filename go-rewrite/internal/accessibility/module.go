package accessibility

import (
	"context"
	"database/sql"
	"net/http"

	"rewritetest/internal/accessibility/internal"
	"rewritetest/internal/accessibility/internal/application"
	"rewritetest/internal/accessibility/internal/domain"
	"rewritetest/internal/accessibility/internal/infrastructure"
	"rewritetest/internal/content"
	"rewritetest/internal/lms"

	"github.com/gin-gonic/gin"
)

type ContentItemService interface {
	GetByCourse(ctx context.Context, courseID int64) ([]content.ContentItem, error)
	CreateManyContentItems(ctx context.Context, contentItems []content.ContentItem) (map[string]int64, error)
}

type ExternalContentRetriever interface {
	GetContent(ctx context.Context, req lms.GetContentRequest) (lms.CourseSyncDataDTO, error)
}

type CourseFileSyncService interface {
	SyncCourseFiles(ctx context.Context, courseID int64, files []lms.FileItemDTO) error
}

type Authenticator interface {
	WithAuth() gin.HandlerFunc
}

type Module struct {
	handler       *internal.Handler
	authenticator internal.Authenticator
	issueRepo     domain.IssueRepository
}

func New(
	db *sql.DB,
	courseRetriever application.CourseRetriever,
	contentItemService ContentItemService,
	externalContentRetriever ExternalContentRetriever,
	courseFileSyncService CourseFileSyncService,
	authenticator Authenticator,
) *Module {
	httpClient := &http.Client{}
	baseURL := "http://udoit3-ace:3000"
	contentScanner := infrastructure.NewEqualAccessScanner(httpClient, baseURL)
	blake3Hasher := infrastructure.NewBlake3ContentHasher()
	reportRepository := infrastructure.NewMySQLReportRepository(db)
	issueRepository := infrastructure.NewMySQLIssueRepository(db)

	scanCourseUseCase := application.NewScanCourseUseCase(
		issueRepository,
		courseRetriever,
		contentItemService,
		externalContentRetriever,
		blake3Hasher,
		courseFileSyncService,
		contentScanner,
	)
	createReportUseCase := application.NewCreateReportUseCase(
		reportRepository,
		issueRepository,
	)
	handler := internal.NewHandler(scanCourseUseCase, createReportUseCase)

	return &Module{
		handler:       handler,
		authenticator: authenticator,
		issueRepo:     issueRepository,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	m.handler.RegisterRoutes(rg, m.authenticator)
}