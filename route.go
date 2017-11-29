package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/code-mobi/kumareport/data"
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
)

func (app *App) IndexHandler(c *gin.Context) {
	var orderCount int64
	app.db.Model(&OrderPayment{}).Count(&orderCount)

	var attendeeTotal int64
	app.db.Model(&Attendee{}).Count(&attendeeTotal)
	var orderTotal int64
	row := app.db.Table("order_payments").Select("SUM(order_total)").Row()
	row.Scan(&orderTotal)

	shirtSizes := GetShirtSizeAmount(app.db)
	skus := GetSkuAmount(app.db)

	var orderPayments []OrderPayment
	app.db.Order("payment_date_time desc").Limit(10).Find(&orderPayments)

	var attendees []Attendee
	app.db.Order("id desc").Limit(10).Find(&attendees)

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

func (app *App) NotFoundHandler(c *gin.Context, message string) {
	c.HTML(http.StatusNotFound, "not_found.tmpl", gin.H{
		"message": message,
	})
}

func (app *App) ReloadDataHandler(c *gin.Context) {
	GenerateOrderPayments(app.db, false)
	GenerateAttendee(app.db, false)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (app *App) OrdersDownloadHandler(c *gin.Context) {
	StreamCSVFile(c, GetOrdersCSV(app.db), GetFileNameNow("orders"))
}

func (app *App) OrderPaymentsDownloadHandler(c *gin.Context) {
	StreamCSVFile(c, GetOrderPaymentsCSV(app.db), GetFileNameNow("order_payment"))
}

func (app *App) OrderPaymentsReloadHandler(c *gin.Context) {
	GenerateOrderPayments(app.db, true)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (app *App) AttendeesHandler(c *gin.Context) {
	var attendees []Attendee
	app.db.Find(&attendees)

	c.HTML(http.StatusOK, "attendees.tmpl", gin.H{
		"attendees": attendees,
	})
}

func (app *App) AttendeesDownloadHandler(c *gin.Context) {
	StreamCSVFile(c, GetAttendeesCSV(app.db), GetFileNameNow("attendee"))
}

func (app *App) AttendeesReloadAllHandler(c *gin.Context) {
	GenerateAttendee(app.db, true)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (app *App) ShirtSizeHandler(c *gin.Context) {

	StreamCSVFile(c, GetShirtSizeCsv(app.db), GetFileNameNow("shirt_size"))
}

func (app *App) ReCheckOnHoldHandler(c *gin.Context) {
	ReCheckOnHold(app.db)
}

func (app *App) LoginHandler(c *gin.Context) {
	type FormLogin struct {
		Login    string `form:"login" json:"user" binding:"required"`
		Password string `form:"password" json:"password" binding:"required"`
	}
	var formLogin FormLogin
	var errorMsg string
	if err := c.ShouldBind(&formLogin); err != nil {
		log.Println("Form Error")
	} else {
		var user data.WpUser
		if formLogin.Login != "" && formLogin.Password != "" {
			app.db.Where("user_login = ? OR user_email = ?", formLogin.Login, formLogin.Login).First(&user)
			user, err = data.UserByLogin(app.db, formLogin.Login)
			if err != nil {
				errorMsg = "Incorrect username or password."
			} else {
				log.Printf("%v", user)
				if data.PasswordHashCheck(formLogin.Password, user.UserPass) {
					errorMsg = "Login Success"
				} else {
					errorMsg = "Incorrect password"
				}
			}
		}
	}

	c.HTML(http.StatusOK, "login.tmpl", gin.H{
		"error": errorMsg,
	})
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
