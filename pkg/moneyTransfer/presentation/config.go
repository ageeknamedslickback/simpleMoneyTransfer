package presentation

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/infrastructure/database/postgresql"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/presentation/middleware"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/presentation/rest"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/usecases"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
)

func Router() *gin.Engine {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v", err)
	}

	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{}))

	db, err := postgresql.ConnectToDatabase()
	if err != nil {
		log.Panicf("server unable to connect to the database: %v", err)
	}
	create := postgresql.NewPostgreSQLDatabase(db)
	get := postgresql.NewPostgreSQLDatabase(db)
	uc := usecases.NewMoneyTransferUsecases(create, get)
	h := rest.NewRestHandlers(uc)

	// Create system accounts
	if err := create.CreateSystemAccount(); err != nil {
		log.Panicf("system error, unable to create default account(s): %v", err)
	}

	gin.DisableConsoleColor()

	f, _ := os.Create("server.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	v1 := router.Group("api/v1")
	v1.POST("/access_token", h.Authenticate)
	v1.Use(adapter.Wrap(middleware.EnsureValidToken()))
	{
		v1.GET("/account/:id", h.Account)
		v1.POST("/account", h.CreateAccount)
		v1.POST("/transfers", h.Transfer)
	}

	return router
}
