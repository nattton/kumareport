package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/gin-gonic/gin"
)

func IndexHandler(c *gin.Context) {
	db, _ := OpenDB()
	defer db.Close()

	var orderCount int64
	db.Model(&OrderPayment{}).Count(&orderCount)

	var attendeeTotal int64
	db.Model(&Attendee{}).Count(&attendeeTotal)
	var orderTotal int64
	row := db.Table("order_payments").Select("SUM(order_total)").Row()
	row.Scan(&orderTotal)

	shirtSizes := GetShirtSizeAmount(db)
	skus := GetSkuAmount(db)

	var orderPayments []OrderPayment
	db.Order("payment_date_time desc").Limit(10).Find(&orderPayments)

	var attendees []Attendee
	db.Order("id desc").Limit(10).Find(&attendees)

	c.HTML(http.StatusOK, "home.tmpl", gin.H{
		"attendeeTotal": fmt.Sprintf("%s", humanize.Comma(attendeeTotal)),
		"orderTotal":    fmt.Sprintf("%s", humanize.Comma(orderTotal)),
		"orderCount":    fmt.Sprintf("%s", humanize.Comma(orderCount)),
		"orderPayments": orderPayments,
		"attendees":     attendees,
		"shirtSizes":    shirtSizes,
		"skus":          skus,
	})
}

func NotFoundHandler(c *gin.Context, message string) {
	c.HTML(http.StatusNotFound, "not_found.tmpl", gin.H{
		"message": message,
	})
}

func ReloadDataHandler(c *gin.Context) {
	db, _ := OpenDB()
	defer db.Close()

	GenerateReportOrderPayment(db)
	GenerateAttendee(db, false)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func OrdersDownloadHandler(c *gin.Context) {
	db, _ := OpenDB()
	defer db.Close()

	StreamCSVFile(c, GetOrdersCSV(db), GetFileNameNow("orders"))
}

func OrderPaymentsDownloadHandler(c *gin.Context) {
	db, _ := OpenDB()
	defer db.Close()

	StreamCSVFile(c, GetOrderPaymentsCSV(db), GetFileNameNow("order_payment"))
}

func OrderPaymentsReloadHandler(c *gin.Context) {
	db, _ := OpenDB()
	defer db.Close()

	GenerateReportOrderPayment(db)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func AttendeesHandler(c *gin.Context) {
	db, _ := OpenDB()
	defer db.Close()

	var attendees []Attendee
	db.Find(&attendees)

	c.HTML(http.StatusOK, "attendees.tmpl", gin.H{
		"attendees": attendees,
	})
}

func AttendeesDownloadHandler(c *gin.Context) {
	db, _ := OpenDB()
	defer db.Close()

	StreamCSVFile(c, GetAttendeesCSV(db), GetFileNameNow("attendee"))
}

func AttendeesReloadAllHandler(c *gin.Context) {
	db, _ := OpenDB()
	defer db.Close()

	GenerateAttendee(db, true)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func ShirtSizeHandler(c *gin.Context) {
	db, _ := OpenDB()
	defer db.Close()

	StreamCSVFile(c, GetShirtSizeCsv(db), GetFileNameNow("shirt_size"))
}

func ReCheckOnHoldHandler(c *gin.Context) {
	db, _ := OpenDB()
	defer db.Close()

	ReCheckOnHold(db)
}

func StreamCSVFile(c *gin.Context, csvData [][]string, fileName string) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	w.WriteAll(csvData) // calls Flush internally
	w.Flush()
	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
	c.Header("Content-Description", "File Transfer")
	fileName = fmt.Sprintf("attachment; filename=%s.csv", fileName)
	c.Header("Content-Disposition", fileName)
	c.Data(http.StatusOK, "text/csv", b.Bytes())
}

func GetFileNameNow(filename string) string {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	t := time.Now().In(loc)
	return fmt.Sprintf("%s-%s", filename, t.Format("2006-01-02_15-04"))
}
