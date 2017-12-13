package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/code-mobi/kumareport/anypay"
	"github.com/code-mobi/kumareport/wp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// const databaseDSN = "root:root@tcp(127.0.0.1:8889)/wordpress?parseTime=true"
const databaseDSN = "mybookzy:ZasSaMi&ZasSaMi@tcp(52.187.124.136:3306)/bookzywpdb?charset=utf8&parseTime=true"
const (
	maxConcurrency = 8
)

var throttle = make(chan int, maxConcurrency)

var cmd string

func init() {
	flag.StringVar(&cmd, "cmd", "", `reload_data
	reload_attendees
	reload_order_payment`)
	flag.Parse()
}

func main() {
	if cmd != "" {
		processCommand(cmd)
		return
	}
	runServer()
}

func processCommand(cmd string) {
	db := OpenDB()
	defer db.Close()
	switch cmd {
	case "reload_data":
		GenerateOrderPayments(db, false)
		GenerateAttendee(db, false)
	case "reload_orders":
		GenerateOrderPayments(db, true)
	case "reload_attendees":
		GenerateAttendee(db, true)
	case "retrieve_products":
		RetrieveProducts(db)
	case "import_stock":
		ImportStock(db)
	case "import_ems":
		ImportEMS(db)
	default:
		log.Println("cmd not found")
	}
}

func ReCheckProcessing(db *gorm.DB) {
	posts := []wp.WpPost{}
	db.Where("post_status = 'wc-processing' AND post_type = 'shop_order'").Find(&posts)
	for _, post := range posts {
		postMeta := GetPostMetaOrder(db, post.ID)
		refID := postMeta["_order_key"]
		status, err := anypay.GetPaymentStatus(refID)

		if err != nil {
			log.Panicf("######### RefID = %s", refID)
		}

		logMessage := fmt.Sprintf("RefID = %s , PaymentCode = %s , PaymentMessage = %s , PaymentType = %s , PaymentDateTime = %s", refID, status.PaymentCode, status.PaymentMessage, status.PaymentType, status.PaymentDateTime)
		if status.PaymentCode != "1" {
			log.Panic("######### " + logMessage)
		} else {
			log.Println(logMessage)
		}
	}
}

func ReCheckOnHold(db *gorm.DB) {
	metaOnholds := []wp.WpPostmeta{}
	db.Where("meta_key = '_date_paid' AND meta_value = ?", "").Find(&metaOnholds)

	for _, meta := range metaOnholds {
		log.Println(meta.PostID)

		metaOrder := wp.WpPostmeta{}
		db.Where("meta_key = '_order_key' AND post_id = ?", meta.PostID).First(&metaOrder)
		checkURL := fmt.Sprintf("https://kumarathonbkk.bookzy.co.th/checkout/order-received/%d/?key=%s", meta.PostID, metaOrder.MetaValue)
		log.Println(checkURL)

		resp, err := http.Get(checkURL)
		if err != nil {
			log.Println(err)
		}
		defer resp.Body.Close()
		log.Println(resp.Status)
	}
}
