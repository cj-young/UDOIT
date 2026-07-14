package lti

import (
	"database/sql"
	"time"

	"rewritetest/internal/lti/internal"
	"rewritetest/internal/lti/internal/application"
	"rewritetest/internal/lti/internal/infrastructure"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Module struct {
	handler *internal.Handler
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	m.handler.RegisterRoutes(rg)
}

func New(client *redis.Client, db *sql.DB, userCreator application.UserCreator, sessionCreator internal.SessionCreator) *Module {
	ltiSessionTTL := 10 * time.Minute

	ltiSessionRepository := infrastructure.NewRedisLTISessionRepository(client, ltiSessionTTL, "lti_session:")
	registrationRepository := infrastructure.NewMySQLRegistrationRepository(db)

	getLaunchRedirectUseCase := application.NewGetLaunchRedirectUseCase(registrationRepository, ltiSessionRepository)
	processLaunchUseCase := application.NewProcessLaunchUseCase(ltiSessionRepository, registrationRepository, infrastructure.NewMySQLLTIUserLinkRepository(db), userCreator)

	handler := internal.NewHandler(sessionCreator, getLaunchRedirectUseCase, processLaunchUseCase)

	return &Module{
		handler: handler,
	}
}
