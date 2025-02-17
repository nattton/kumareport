package main

import (
	"strconv"
	"strings"

	"github.com/code-mobi/kumareport/wp"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

func OpenRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return client
}

func Float64ToString(i float64) string {
	v := strconv.FormatFloat(i, 'f', 2, 64)
	return v
}

func GetPostMetaOrder(db *gorm.DB, postID int) map[string]string {
	metaKeys := []string{"_order_key", "_payment_method", "_paid_date", "_shipping_first_name", "_shipping_last_name", "_billing_phone", "_billing_email", "_shipping_address_index", "_order_total"}
	postMeta := wp.GetPostMetaFields(db, postID, metaKeys)

	return postMeta
}

func GetPostMetaAttendee(db *gorm.DB, postID int) *Attendee {
	metaKeys := []string{kTicketTypeID, kTicketCode, kFirstname, kLastname, kPhone, kGender, kBirthday, kEmail, kIDCard, kAddress}
	postMeta := wp.GetPostMetaFields(db, postID, metaKeys)
	orderID, _ := strconv.Atoi(strings.Split(postMeta[kTicketCode], "-")[0])
	ticketTypeID, _ := strconv.Atoi(postMeta[kTicketTypeID])
	attendee := &Attendee{
		ID:           postID,
		OrderID:      orderID,
		TicketTypeID: ticketTypeID,
		TicketCode:   postMeta[kTicketCode],
		Firstname:    postMeta[kFirstname],
		Lastname:     postMeta[kLastname],
		Phone:        postMeta[kPhone],
		Gender:       postMeta[kGender],
		Birthday:     postMeta[kBirthday],
		Email:        postMeta[kEmail],
		IDCard:       postMeta[kIDCard],
		Address:      postMeta[kAddress],
	}
	return attendee
}
