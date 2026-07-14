package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"rewritetest/internal/auth"
	"rewritetest/internal/files"
	"rewritetest/internal/i18n"
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
	authModule := auth.New(client, db)
	filesModule := files.New(db)
	i18nModule := i18n.New()
	usersModule := users.New(db, i18nModule, authModule)
	ltiModule := lti.New(client, db, usersModule, authModule)

	// Register routes
	usersModule.RegisterRoutes(router.Group("/users"))
	ltiModule.RegisterRoutes(router.Group("/lti"))
	filesModule.RegisterRoutes(router.Group("/files"))

	port := ":8080"
	fmt.Println("Server running on http://localhost" + port)

	err = router.Run(port)
	if err != nil {
		log.Fatal(err)
	}
}
