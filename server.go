package main

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func runServer() {
	router := gin.New()
	router.Use(gin.Logger())

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	router.LoadHTMLGlob(dir + "/templates/*")

	authorized := router.Group("/")
	authorized.Use(gin.BasicAuth(gin.Accounts{
		"kumamon": "kumakuma555",
	}))
	{
		authorized.GET("/", IndexHandler)
		authorized.GET("/reload_data", ReloadDataHandler)
		authorized.GET("/orders", OrdersHandler)
		authorized.GET("/order/:id", OrderHandler)
		authorized.GET("/orders/download", OrdersDownloadHandler)
		authorized.GET("/order_payments/reload", OrderPaymentsReloadHandler)
		authorized.GET("/order_payments/download", OrderPaymentsDownloadHandler)

		authorized.GET("/attendees", AttendeesHandler)
		authorized.GET("/attendees/download", AttendeesDownloadHandler)
		authorized.GET("/attendees/reload", AttendeesReloadAllHandler)
		authorized.GET("/attendee/:id", AttendeeHandler)
		authorized.POST("/attendee/:id", AttendeeHandler)
		authorized.POST("/attendee/:id/edit", AttendeeUpdateHandler)

		authorized.GET("/shirtsizes/download", ShirtSizeHandler)

		authorized.GET("/recheck_onhold", ReCheckOnHoldHandler)

	}

	api := router.Group("/api")
	{
		api.GET("/attendees", ApiAttendeesHandler)
	}

	router.GET("/login", LoginHandler)
	router.POST("/login", LoginHandler)

	router.Run(":3000")
}
