package main

import (
	"errors"
	"net/http"
	"strconv"

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
	PaymentStatus PaymentStatus
	OrderItems    []OrderItem
	Attendees     []Attendee
}

type OrderItem struct {
	ItemName  string
	Qty       string
	LineTotal string
}

func OrdersHandler(c *gin.Context) {
	q := c.Query("q")

	db, _ := OpenDB()
	defer db.Close()

	orders := []Order{}
	if q != "" {
		rows, err := db.Raw("SELECT post_id FROM wp_postmeta WHERE post_id = ? OR (meta_key IN('_shipping_first_name', '_shipping_last_name', '_billing_phone') AND meta_value = ?) GROUP BY post_id", q, q).Rows()
		defer rows.Close()

		for rows.Next() {
			var orderID int
			rows.Scan(&orderID)
			var post WpPost
			db.First(&post, orderID)
			postMeta := getPostMeta(db, orderID)
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
			NotFoundHandler(c, err.Error())
			return
		}

		c.HTML(http.StatusOK, "orders.tmpl", gin.H{
			"orders": orders,
		})
	} else {
		var orderPayments []OrderPayment
		db.Order("payment_date_time desc").Find(&orderPayments)
		c.HTML(http.StatusOK, "order_payments.tmpl", gin.H{
			"orderPayments": orderPayments,
		})
	}
}

func OrderHandler(c *gin.Context) {
	id := c.Param("id")
	orderID, err := strconv.Atoi(id)

	db, _ := OpenDB()
	defer db.Close()

	if err != nil {
		c.HTML(http.StatusNotFound, "order.tmpl", gin.H{})
		return
	}

	order, err := GetOrder(db, orderID)
	if err != nil {
		NotFoundHandler(c, err.Error())
		return
	}

	c.HTML(http.StatusOK, "order.tmpl", gin.H{
		"order": order,
	})
}

func OrderDownloadHandler(c *gin.Context) {
	db, _ := OpenDB()
	defer db.Close()

	StreamCSVFile(c, GetOrderReport(db), GetFileNameNow("order"))
}

func ApiOrderHandler(c *gin.Context) {
	id := c.Query("id")
	orderID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID Invalid"})
		return
	}
	db, _ := OpenDB()
	defer db.Close()

	order, err := GetOrder(db, orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

func GetOrder(db *gorm.DB, orderID int) (Order, error) {
	order := Order{}
	var post WpPost
	db.First(&post, orderID)
	orderMeta := getPostMetaFields(db, orderID, []string{"_order_key", "_payment_method", "_order_total", "_shipping_first_name", "_shipping_last_name", "_billing_phone", "_billing_email"})
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
	order.PaymentStatus = GetPaymentStatus(order.RefID)
	order.OrderItems = GetOrderItem(db, order.OrderID)
	order.Attendees = GetAttendees(db, order.OrderID)
	return order, nil
}

func GetOrderItem(db *gorm.DB, id int) []OrderItem {
	orderItems := []OrderItem{}
	wpOrderItems := []WpWoocommerceOrderItem{}
	db.Where("order_item_type = 'line_item' AND order_id = ?", id).Find(&wpOrderItems)
	for _, wpOrderItem := range wpOrderItems {
		orderItem := OrderItem{}
		orderItemmeta := getOrderItemmeta(db, wpOrderItem.OrderItemID)
		orderItem.ItemName = wpOrderItem.OrderItemName
		orderItem.Qty = orderItemmeta["_qty"]
		orderItem.LineTotal = orderItemmeta["_line_total"]
		orderItems = append(orderItems, orderItem)
	}
	return orderItems
}
