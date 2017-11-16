package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

const authKey = "bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1bmlxdWVfbmFtZSI6ImY3MTgyM2UzLWJhOGMtNGMwNi05ODkzLTA0MDgzMzJhNDA1NSIsInN1YiI6ImJvb2t6eUBhbnlwYXkuY28udGgiLCJlbWFpbCI6ImJvb2t6eUBhbnlwYXkuY28udGgiLCJyb2xlIjoiQU5ZUEFfV0VCIiwiaXNzIjoiYXV0aC5hbnlwYXkuY28udGgiLCJhdWQiOiI2ODViZDg0MmY3ZDQ0NmU2OWY4Yjc4ZTY0MzJjNzY3OCIsImV4cCI6MTUxNTUzNDYxNCwibmJmIjoxNTEwMzUwNjE0fQ.M7CB3T2PmzT2L9laT2m-8WiJpHCD1RcqEvv0SoepT6Q"

// const databaseDSN = "root:root@tcp(127.0.0.1:8889)/wordpress?parseTime=true"

const databaseDSN = "mybookzy:ZasSaMi&ZasSaMi@tcp(52.187.124.136:3306)/bookzywpdb?charset=utf8&parseTime=true"
const (
	maxConcurrency = 8
)

var throttle = make(chan int, maxConcurrency)

var cmd string

func init() {
	flag.StringVar(&cmd, "cmd", "", `reload_attendees
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
	db, _ := OpenDB()
	defer db.Close()
	switch cmd {
	case "reload_orders":
		GenerateReportOrderPayment(db)
	case "reload_attendees":
		GenerateAttendee(db, true)
	case "reload_order_payment":
		GenerateReportOrderPayment(db)
	case "import_shirt":
		ImportShirt(db)
	case "import_change_shirt":
		ImportChangeShirt(db)
	case "retrieve_product":
		RetrieveInventory(db)
	case "import_stock":
		ImportStock(db)
	case "import_stock_left":
		ImportStockLeft(db)
	default:
		log.Println("cmd not found")
	}
}

func ReCheckProcessing(db *gorm.DB) {
	posts := []WpPost{}
	db.Where("post_status = 'wc-processing' AND post_type = 'shop_order'").Find(&posts)
	for _, post := range posts {
		postMeta := getPostMeta(db, post.ID)
		refID := postMeta["_order_key"]
		status := GetPaymentStatus(refID)

		if status.TotalRow != 1 {
			log.Panicf("######### RefID = %s , TotalRow = %d", refID, status.TotalRow)
		}

		statusRow := status.DataRow[0]

		logMessage := fmt.Sprintf("RefID = %s , PaymentCode = %s , PaymentMessage = %s , PaymentType = %s , PaymentDateTime = %s", refID, statusRow.PaymentCode, statusRow.PaymentMessage, statusRow.PaymentType, statusRow.PaymentDateTime)
		if statusRow.PaymentCode != "1" {
			log.Panic("######### " + logMessage)
		} else {
			log.Println(logMessage)
		}
	}
}

func ReCheckOnHold(db *gorm.DB) {
	metaOnholds := []WpPostmeta{}
	db.Where("meta_key = '_date_paid' AND meta_value = ?", "").Find(&metaOnholds)

	for _, meta := range metaOnholds {
		log.Println(meta.PostID)

		metaOrder := WpPostmeta{}
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

func GetPaymentStatus(refID string) PaymentStatus {
	body := callAnypay(fmt.Sprintf("/payment/status2?RefId=%s", refID))
	var status PaymentStatus
	err := json.Unmarshal(body, &status)
	if err != nil {
		log.Fatal(err)
	}
	return status
}

func callAnypay(apiPath string) []byte {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.anypay.co.th%s", apiPath), nil)
	req.Header.Set("Authorization", authKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	log.Println(resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	return body
}
