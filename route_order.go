package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/code-mobi/kumareport/anypay"
	"github.com/code-mobi/kumareport/wp"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Order struct {
	OrderID       int
	Status        string
	RefID         string
	PaymentMethod string
	OrderTotal    float64
	Firstname     string
	Lastname      string
	Phone         string
	Email         string
	PaymentStatus anypay.PaymentStatus
	OrderItems    OrderItems
	Attendees     Attendees
}

type OrderItem struct {
	ID        int
	Name      string
	Type      string
	Qty       string
	LineTotal float64
}

type OrderItems []*OrderItem

func (app *App) OrdersHandler(c *gin.Context) {
	q := c.Query("q")

	orders := []Order{}
	if q != "" {
		rows, err := app.db.Raw("SELECT post_id FROM wp_postmeta WHERE post_id = ? OR (meta_key IN('_shipping_first_name', '_shipping_last_name', '_billing_phone') AND meta_value = ?) GROUP BY post_id", q, q).Rows()
		defer rows.Close()

		for rows.Next() {
			var orderID int
			rows.Scan(&orderID)
			var post wp.WpPost
			app.db.First(&post, orderID)
			postMeta := GetPostMetaOrder(app.db, orderID)
			orderTotal, _ := strconv.ParseFloat(postMeta["_order_total"], 64)
			order := Order{
				OrderID:    orderID,
				Status:     post.PostStatus,
				RefID:      postMeta["_order_key"],
				Firstname:  postMeta["_shipping_first_name"],
				Lastname:   postMeta["_shipping_last_name"],
				Phone:      postMeta["_billing_phone"],
				OrderTotal: orderTotal,
			}
			orders = append(orders, order)
		}
		if err != nil {
			app.NotFoundHandler(c, err.Error())
			return
		}

		c.HTML(http.StatusOK, "orders.tmpl", gin.H{
			"orders": orders,
		})
	} else {
		var orderPayments []OrderPayment
		app.db.Order("payment_date_time desc").Find(&orderPayments)
		c.HTML(http.StatusOK, "order_payments.tmpl", gin.H{
			"orderPayments": orderPayments,
		})
	}
}

func (app *App) OrderHandler(c *gin.Context) {
	id := c.Param("id")
	orderID, err := strconv.Atoi(id)

	if err != nil {
		c.HTML(http.StatusNotFound, "order.tmpl", gin.H{})
		return
	}

	order, err := GetOrder(app.db, orderID)
	if err != nil {
		app.NotFoundHandler(c, err.Error())
		return
	}

	c.HTML(http.StatusOK, "order.tmpl", gin.H{
		"order": order,
	})
}

func GetOrder(db *gorm.DB, orderID int) (Order, error) {
	order := Order{}
	var post wp.WpPost
	db.First(&post, orderID)
	orderMeta := wp.GetPostMetaFields(db, orderID, []string{"_order_key", "_payment_method", "_order_total", "_shipping_first_name", "_shipping_last_name", "_billing_phone", "_billing_email"})
	if orderMeta["_order_key"] == "" {
		return order, errors.New("Order Not Found")
	}
	order.OrderID = orderID
	order.Status = post.PostStatus
	order.RefID = orderMeta["_order_key"]
	orderTotal, _ := strconv.ParseFloat(orderMeta["_order_total"], 64)
	order.OrderTotal = orderTotal
	order.PaymentMethod = orderMeta["_payment_method"]
	order.Firstname = orderMeta["_shipping_first_name"]
	order.Lastname = orderMeta["_shipping_last_name"]
	order.Phone = orderMeta["_billing_phone"]
	order.Email = orderMeta["_billing_email"]
	order.PaymentStatus, _ = anypay.GetPaymentStatus(order.RefID)
	order.OrderItems = GetOrderItem(db, order.OrderID)
	order.Attendees = GetAttendees(db, order.OrderID)
	return order, nil
}
