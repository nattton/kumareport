package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func OpenDB() (*gorm.DB, error) {
	db, err := gorm.Open("mysql", databaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(gin.Mode() == "debug")
	return db, err
}

type WpPost struct {
	ID         int `gorm:"primary_key;column:ID"`
	PostDate   string
	PostTitle  string
	PostStatus string
	PostType   string
}

type WpPostmeta struct {
	MetaID    int `gorm:"primary_key"`
	PostID    int
	MetaKey   string
	MetaValue string
}

type PaymentStatus struct {
	TotalRow int
	DataRow  []StatusDataRow
}

type StatusDataRow struct {
	RefID           string
	TransDesc       string
	TransId         string
	PaymentCode     string
	PaymentMessage  string
	Amount          float64
	PaymentType     string
	PaymentAmount   string
	PaymentDateTime string
	OrderNo         string
}

type WpWoocommerceOrderItem struct {
	OrderItemID   int
	OrderItemName string
	OrderItemType string
	OrderID       int
}

type WpWoocommerceOrderItemmeta struct {
	MetaID      int
	OrderItemID int
	MetaKey     string
	MetaValue   string
}

func getPostMeta(db *gorm.DB, postID int) map[string]string {
	postMeta := make(map[string]string)
	postMetas := []WpPostmeta{}
	metaKeys := []string{"_order_key", "_payment_method", "_paid_date", "_shipping_first_name", "_shipping_last_name", "_billing_phone", "_billing_email", "_shipping_address_index", "_order_total"}
	db.Where("post_id = ? AND meta_key IN (?)", postID, metaKeys).Find(&postMetas)

	for i := range postMetas {
		postMeta[postMetas[i].MetaKey] = postMetas[i].MetaValue
	}
	return postMeta
}

func getPostMetaFields(db *gorm.DB, postID int, metaKeys []string) map[string]string {
	postMeta := make(map[string]string)
	postMetas := []WpPostmeta{}
	db.Where("post_id = ? AND meta_key IN (?)", postID, metaKeys).Find(&postMetas)

	for i := range postMetas {
		postMeta[postMetas[i].MetaKey] = postMetas[i].MetaValue
	}
	return postMeta
}

func getPostMetaAttendee(db *gorm.DB, postID int) Attendee {
	metaKeys := []string{kTicketTypeID, kTicketCode, kFirstname, kLastname, kPhone, kGender, kBirthday, kEmail, kIDCard, kAddress}
	postMeta := getPostMetaFields(db, postID, metaKeys)
	orderID, _ := strconv.Atoi(strings.Split(postMeta[kTicketCode], "-")[0])
	ticketTypeID, _ := strconv.Atoi(postMeta[kTicketTypeID])
	attendee := Attendee{
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

func getOrderItemmeta(db *gorm.DB, orderItemID int) map[string]string {
	postMeta := make(map[string]string)
	postMetas := []WpWoocommerceOrderItemmeta{}
	metaKeys := []string{"_qty", "_line_total"}
	db.Where("order_item_id = ? AND meta_key IN (?)", orderItemID, metaKeys).Find(&postMetas)

	for i := range postMetas {
		postMeta[postMetas[i].MetaKey] = postMetas[i].MetaValue
	}
	return postMeta
}

func Float64ToString(i float64) string {
	v := strconv.FormatFloat(i, 'f', 2, 64)
	return v
}
