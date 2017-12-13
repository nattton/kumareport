package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type App struct {
	db *gorm.DB
}

func runServer() {
	db := OpenDB()
	defer db.Close()
	app := &App{db}
	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob(htmlDir + "/*")

	authorized := router.Group("/")
	authorized.Use(gin.BasicAuth(gin.Accounts{
		"kumamon": "kumakuma555",
	}))
	{
		authorized.GET("/", app.IndexHandler)
		authorized.GET("/reload_data", app.ReloadDataHandler)
		authorized.GET("/orders", app.OrdersHandler)
		authorized.GET("/order/:id", app.OrderHandler)
		authorized.GET("/orders/download", app.OrdersDownloadHandler)
		authorized.GET("/order_payments/reload", app.OrderPaymentsReloadHandler)
		authorized.GET("/order_payments/download", app.OrderPaymentsDownloadHandler)

		authorized.GET("/attendees", app.AttendeesHandler)
		authorized.GET("/attendees/download", app.AttendeesDownloadHandler)
		authorized.GET("/attendees/reload", app.AttendeesReloadAllHandler)
		authorized.GET("/attendee/:id", app.AttendeeHandler)
		authorized.POST("/attendee/:id", app.AttendeeHandler)
		authorized.POST("/attendee/:id/edit", app.AttendeeUpdateHandler)

		authorized.GET("/shirtsizes/download", app.ShirtSizeHandler)

		authorized.GET("/recheck_onhold", app.ReCheckOnHoldHandler)

	}

	api := router.Group("/api")
	{
		api.GET("/attendees", app.ApiAttendeesHandler)
	}

	router.GET("/login", app.LoginHandler)
	router.POST("/login", app.LoginHandler)

	router.Run(":3000")
}

func OpenDB() *gorm.DB {
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(gin.Mode() == "debug")
	db.AutoMigrate(&Attendee{}, &Product{}, &OrderPayment{})
	return db
}
