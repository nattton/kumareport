package main

import "github.com/gin-gonic/gin"

func runServer() {
	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*")

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

	router.Run(":3000")
}
