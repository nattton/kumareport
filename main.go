package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

const authKey = "bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1bmlxdWVfbmFtZSI6ImY3MTgyM2UzLWJhOGMtNGMwNi05ODkzLTA0MDgzMzJhNDA1NSIsInN1YiI6ImJvb2t6eUBhbnlwYXkuY28udGgiLCJlbWFpbCI6ImJvb2t6eUBhbnlwYXkuY28udGgiLCJyb2xlIjoiQU5ZUEFfV0VCIiwiaXNzIjoiYXV0aC5hbnlwYXkuY28udGgiLCJhdWQiOiI2ODViZDg0MmY3ZDQ0NmU2OWY4Yjc4ZTY0MzJjNzY3OCIsImV4cCI6MTUxNTUzNDYxNCwibmJmIjoxNTEwMzUwNjE0fQ.M7CB3T2PmzT2L9laT2m-8WiJpHCD1RcqEvv0SoepT6Q"

// const databaseDSN = "root:root@tcp(127.0.0.1:8889)/wordpress?parseTime=true"

const databaseDSN = "mybookzy:ZasSaMi&ZasSaMi@tcp(52.187.124.136:3306)/bookzywpdb?charset=utf8&parseTime=true"
const (
	limitRow       = 40
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
		authorized.GET("/orders", OrdersHandler)
		authorized.GET("/order/:id", OrderHandler)
		authorized.GET("/orders/download", OrdersDownloadHandler)
		authorized.GET("/order_payments/reload", OrderPaymentsReloadHandler)
		authorized.GET("/order_payments/download", OrderPaymentsDownloadHandler)

		// authorized.GET("/orders/download", OrderDownloadHandler)

		authorized.GET("/attendees", AttendeesHandler)
		authorized.GET("/attendees/download", AttendeesDownloadHandler)
		authorized.GET("/attendees/reload", AttendeesReloadHandler)
		authorized.GET("/attendees/reload_all", AttendeesReloadAllHandler)
		authorized.GET("/attendee/:id", AttendeeHandler)
		authorized.POST("/attendee/:id", AttendeeHandler)
		authorized.POST("/attendee/:id/edit", AttendeeUpdateHandler)

		authorized.GET("/shirtsizes/download", ShirtSizeHandler)
		authorized.GET("/generate_order", GenerateOrderHandler)
		authorized.GET("/recheck_onhold", ReCheckOnHoldHandler)

	}

	routerApi := router.Group("/api")
	{
		routerApi.GET("/order", ApiOrderHandler)
	}

	router.Run(":3000")
}

func reCheckProcessing(db *gorm.DB) {
	posts := []WpPost{}
	db.Where("post_status = 'wc-processing' AND post_type = 'shop_order'").Find(&posts)

	// log.Printf("wc-processing count = %d", len(posts))

	csvStatus := [][]string{{"Order ID", "RefID", "PaymentType", "PaymentAmount", "PaymentDateTime", "Firstname", "Lastname", "Phone", "OrderTotal"}}

	for _, post := range posts {
		// metaOrder := WpPostmeta{}
		// db.Where("meta_key = '_order_key' AND post_id = ?", meta.PostID).First(&metaOrder)
		// refID := metaOrder.MetaValue
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

			csvStatus = append(csvStatus, []string{
				strconv.Itoa(post.ID),
				statusRow.RefID,
				statusRow.PaymentType,
				statusRow.PaymentAmount,
				statusRow.PaymentDateTime,
				postMeta["_shipping_first_name"],
				postMeta["_shipping_last_name"],
				postMeta["_billing_phone"],
				postMeta["_order_total"],
			})
		}
	}

	csvfile, err := os.Create("kumamon_success.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer csvfile.Close()

	w := csv.NewWriter(csvfile)
	w.WriteAll(csvStatus) // calls Flush internally

	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}

}

func reCheckPendingPayment(db *gorm.DB) {
	metaOnholds := []WpPostmeta{}
	db.Where("meta_key = '_date_paid' AND meta_value = ?", "").Find(&metaOnholds)

	for _, meta := range metaOnholds {
		log.Println(meta.PostID)

		metaOrder := WpPostmeta{}
		db.Where("meta_key = '_order_stock_reduced' AND meta_value = 'yes' AND post_id = ?", meta.PostID).First(&metaOrder)

		if metaOrder.MetaValue != "" {
			continue
		}

		db.Where("meta_key = '_order_key' AND post_id = ?", meta.PostID).First(&metaOrder)
		refID := metaOrder.MetaValue
		status := GetPaymentStatus(refID)

		if status.TotalRow != 1 {
			log.Printf("######### RefID = %s , TotalRow = %d", refID, status.TotalRow)
			continue
		}

		statusRow := status.DataRow[0]

		logMessage := fmt.Sprintf("RefID = %s , PaymentCode = %s , PaymentMessage = %s , PaymentType = %s", refID, statusRow.PaymentCode, statusRow.PaymentMessage, statusRow.PaymentType)
		if statusRow.PaymentCode == "1" {
			log.Panic("######### " + logMessage)
		} else {
			log.Println(logMessage)
		}

	}
}

func reCheckOnHold(db *gorm.DB) {
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

func ImportShirt(db *gorm.DB) {
	m := NewModel(db)
	m.RetieveTickets()
	m.RetieveSkuList()
	file, err := os.Open("attendee-2017-11-15_15-52.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()
	//get a new cvsReader for reading file
	reader := csv.NewReader(file)
	var lineCount int
	for {

		// read just one record, but we could ReadAll() as well
		record, err := reader.Read()

		// end-of-file is fitted into err
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			reader.Read()
			continue
		}
		// record is array of strings Ref http://golang.org/src/pkg/encoding/csv/reader.go?s=#L134
		if lineCount > 0 {
			fmt.Printf("Record %s %s\n", record[0], record[8])
			id, err := strconv.Atoi(record[0])
			if err != nil {
				continue
			}
			product := m.GetSkuList(record[8])
			db.Exec("UPDATE wp_postmeta SET meta_value= ? WHERE meta_key='ticket_type_id' AND post_id = ?", product.ID, record[0])
			attendee := GetAttendee(db, m, id)
			db.Save(&attendee)
		}
		lineCount++
	}
}

func ImportChangeShirt(db *gorm.DB) {
	m := NewModel(db)
	m.RetieveTickets()
	m.RetieveSkuList()
	file, err := os.Open("ChangeSize2.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()
	//get a new cvsReader for reading file
	reader := csv.NewReader(file)
	var lineCount int
	for {

		// read just one record, but we could ReadAll() as well
		record, err := reader.Read()

		// end-of-file is fitted into err
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			reader.Read()
			continue
		}
		// record is array of strings Ref http://golang.org/src/pkg/encoding/csv/reader.go?s=#L134
		if lineCount > 0 {
			fmt.Printf("Record %s %s\n", record[0], record[1], record[2])
			postMetas := []WpPostmeta{}
			db.Where("meta_value LIKE ? AND meta_key = ?", "%"+record[2]+"%", kPhone).Find(&postMetas)
			resultCount := len(postMetas)
			if resultCount != 1 {
				log.Panicf("Result =%d | MObile No. %s", resultCount, record[2])
			}
			product := m.GetSkuList(record[1])

			db.Exec("UPDATE wp_postmeta SET meta_value= ? WHERE meta_key='ticket_type_id' AND post_id = ?", product.ID, postMetas[0].PostID)
			attendee := GetAttendee(db, m, postMetas[0].PostID)
			db.Save(&attendee)
			// break
		}
		lineCount++
	}
}

func ImportStock(db *gorm.DB) {
	lines := readCSV("KumamonInventory.csv")
	for _, line := range lines {
		db.Exec("UPDATE products SET stock= ? WHERE sku=?", line[1], line[0])
	}
}

func ImportStockLeft(db *gorm.DB) {
	lines := readCSV("KumamonInventoryLeft.csv")
	stockLeft := make(map[string]int)
	for _, line := range lines {
		left, _ := strconv.Atoi(line[1])
		stockLeft[line[0]] = left
	}

	skus := GetSkuAmount(db)
	for _, sku := range skus {
		stock := stockLeft[sku.ShirtSize] + sku.Amount
		db.Exec("UPDATE products SET stock= ? WHERE sku=?", stock, sku.ShirtSize)
	}
}

func readFile(file string) []byte {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return b
}

func readCSV(file string) [][]string {
	b := readFile(file)
	r := csv.NewReader(bytes.NewReader(b))
	r.Comma = ','
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	return records
}
