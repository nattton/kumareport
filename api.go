package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

const kCacheApiAttendees = "api_attendee"

func (app *App) ApiAttendeesHandler(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*.bookzy.co.th")
	type Attendee struct {
		ID        int
		OrderID   int
		Firstname string
		Lastname  string
		ItemName  string
		EMS       string
	}
	var attendees []Attendee
	redisClient := OpenRedis()
	results, err := redisClient.Get(kCacheApiAttendees).Result()
	if err != nil || err == redis.Nil {
		app.db.Raw("SELECT id, order_id, firstname, lastname, sku item_name, ems FROM attendees ORDER BY order_id, id").Scan(&attendees)
		c.JSON(http.StatusOK, attendees)

		b, err := json.Marshal(attendees)
		if err != nil {
			log.Println(err)
		}

		err = redisClient.Set(kCacheApiAttendees, b, 5*time.Minute).Err()
		if err != nil {
			log.Println(err)
		}
		return
	}

	json.Unmarshal([]byte(results), &attendees)
	c.JSON(http.StatusOK, attendees)
}
