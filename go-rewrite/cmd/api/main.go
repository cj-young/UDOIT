package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"rewritetest/internal/auth"
	"rewritetest/internal/courses"
	"rewritetest/internal/files"
	"rewritetest/internal/i18n"
	"rewritetest/internal/lms"
	"rewritetest/internal/lti"
	"rewritetest/internal/shared/http/middleware"
	"rewritetest/internal/users"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello!")
}

func main() {
	router := gin.Default()

	router.Use(middleware.ErrorHandler())

	// Initialize infrastructure
	db, err := sql.Open("mysql", os.Getenv("GO_DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("MySQL database successfully connected")

	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("GO_REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	err = client.Ping(context.Background()).Err()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Redis client successfully connected")

	// Initialize modules
	lmsModule := lms.New(db)
	authModule := auth.New(client, db)
	coursesModule := courses.New(db)
	filesModule := files.New(db, lmsModule)
	i18nModule := i18n.New()
	usersModule := users.New(db, i18nModule, authModule)
	ltiModule := lti.New(client, db, usersModule, authModule, coursesModule, os.Getenv("GO_BASE_URL"))

	// Register routes
	usersModule.RegisterRoutes(router.Group("/users"))
	ltiModule.RegisterRoutes(router.Group("/lti"))
	filesModule.RegisterRoutes(router.Group("/files"))

	mockToolDir := filepath.Clean("./mock-lti-tool")
	serveMockTool := func(c *gin.Context) {
		requestPath := strings.TrimPrefix(c.Param("path"), "/")
		if requestPath == "" {
			c.File(filepath.Join(mockToolDir, "index.html"))
			return
		}

		candidate := filepath.Clean(filepath.Join(mockToolDir, requestPath))
		if !strings.HasPrefix(candidate, mockToolDir+string(filepath.Separator)) {
			c.File(filepath.Join(mockToolDir, "index.html"))
			return
		}

		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			c.File(candidate)
			return
		}

		c.File(filepath.Join(mockToolDir, "index.html"))
	}

	router.GET("/mock-lti-tool", func(c *gin.Context) {
		slog.Info("Serving mock LTI tool index page")
		c.File(filepath.Join(mockToolDir, "index.html"))
	})
	router.GET("/mock-lti-tool/*path", serveMockTool)
	router.POST("/mock-lti-tool/*path", serveMockTool)

	port := ":8080"
	fmt.Println("Server running on http://localhost" + port)

	err = router.Run(port)
	if err != nil {
		log.Fatal(err)
	}
}
