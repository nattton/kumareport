package main

import (
	"fmt"
	"strconv"

	"github.com/code-mobi/kumareport/wp"
	"github.com/jinzhu/gorm"
)

func GenerateOrderPayments(db *gorm.DB, forceUpdate bool) {
	posts := wp.WpPosts{}
	db.Where("post_status = 'wc-processing' AND post_type = 'shop_order'").Find(&posts)
	m := NewModel(db)
	if len(posts) > 0 {
		m.RetieveOrderPayment()
	}
	for _, post := range posts {
		orderPayment := m.GetOrderPayment(post.ID)
		if orderPayment.OrderID != post.ID {
			orderPayment := GetPostMetaOrderPayment(db, post.ID)
			db.Create(&orderPayment)
		} else if forceUpdate {
			orderPayment := GetPostMetaOrderPayment(db, post.ID)
			db.Save(&orderPayment)
		}
	}
}

func GetOrderPaymentsCSV(db *gorm.DB) [][]string {
	orderPayments := OrderPayments{}
	db.Order("payment_date_time").Find(&orderPayments)
	csvData := [][]string{{
		"OrderID",
		"RefID",
		"PaymentType",
		"PaymentAmount",
		"PaymentDateTime",
		"Firstname",
		"Lastname",
		"Phone",
		"OrderTotal",
		"CreditCardFee",
		"VatFee",
		"TaxFee",
		"OrderCreditCardFee",
		"OrderVatFee",
		"OrderTaxFee",
		"OrderAfterVat",
		"OrderAfterFee",
	}}
	for _, order := range orderPayments {
		csvData = append(csvData, []string{
			strconv.Itoa(order.OrderID),
			order.RefID,
			order.PaymentType,
			Float64ToString(order.PaymentAmount),
			order.PaymentDateTime,
			order.Firstname,
			order.Lastname,
			order.Phone,
			Float64ToString(order.OrderTotal),
			Float64ToString(order.CreditCardFee),
			Float64ToString(order.VatFee),
			Float64ToString(order.TaxFee),
			Float64ToString(order.OrderCreditCardFee),
			Float64ToString(order.OrderVatFee),
			Float64ToString(order.OrderTaxFee),
			Float64ToString(order.OrderAfterVat),
			Float64ToString(order.OrderAfterFee),
		})
	}
	return csvData
}

func GetOrdersCSV(db *gorm.DB) [][]string {
	orderPayments := OrderPayments{}
	db.Order("order_id").Find(&orderPayments)
	csvData := [][]string{{
		"OrderID",
		"Firstname",
		"Lastname",
		"Phone",
		"OrderTotal",
		"Company",
		"Address",
		"Postcode",
		"Attendee ID",
		"Sku",
		"Attendee Firstname",
		"Attendee Lastname",
		"Attendee Phone",
		"Attendee Gender",
		"Attendee Birthday",
		"Attendee Email",
		"Attendee IDCard",
		"Attendee Address",
	}}

	type ReportOrder struct {
		OrderPayment OrderPayment
		Attendee     Attendee
	}

	rows, _ := db.Raw("SELECT o.order_id, o.firstname, o.lastname, o.phone, o.order_total, o.shipping_company, o.shipping_address, o.shipping_postcode, a.id, a.firstname, a.lastname, a.phone, a.gender, a.birthday, a.email, a.id_card, a.address, a.sku FROM order_payments o INNER JOIN attendees a ON (o.order_id = a.order_id) ORDER BY o.order_id").Rows()
	defer rows.Close()
	var firstLineID int
	for rows.Next() {
		report := ReportOrder{OrderPayment{}, Attendee{}}

		rows.Scan(&report.OrderPayment.OrderID,
			&report.OrderPayment.Firstname,
			&report.OrderPayment.Lastname,
			&report.OrderPayment.Phone,
			&report.OrderPayment.OrderTotal,
			&report.OrderPayment.ShippingCompany,
			&report.OrderPayment.ShippingAddress,
			&report.OrderPayment.ShippingPostcode,
			&report.Attendee.ID,
			&report.Attendee.Firstname,
			&report.Attendee.Lastname,
			&report.Attendee.Phone,
			&report.Attendee.Gender,
			&report.Attendee.Birthday,
			&report.Attendee.Email,
			&report.Attendee.IDCard,
			&report.Attendee.Address,
			&report.Attendee.Sku,
		)
		if firstLineID != report.OrderPayment.OrderID {
			csvData = append(csvData, []string{
				strconv.Itoa(report.OrderPayment.OrderID),
				report.OrderPayment.Firstname,
				report.OrderPayment.Lastname,
				report.OrderPayment.Phone,
				Float64ToString(report.OrderPayment.OrderTotal),
				report.OrderPayment.ShippingCompany,
				report.OrderPayment.ShippingAddress,
				report.OrderPayment.ShippingPostcode,
				strconv.Itoa(report.Attendee.ID),
				report.Attendee.Sku,
				report.Attendee.Firstname,
				report.Attendee.Lastname,
				report.Attendee.Phone,
				report.Attendee.Gender,
				report.Attendee.Birthday,
				report.Attendee.Email,
				report.Attendee.IDCard,
				report.Attendee.Address,
			})
		} else {
			csvData = append(csvData, []string{
				"", "", "", "", "", "", "", "",
				strconv.Itoa(report.Attendee.ID),
				report.Attendee.Sku,
				report.Attendee.Firstname,
				report.Attendee.Lastname,
				report.Attendee.Phone,
				report.Attendee.Gender,
				report.Attendee.Birthday,
				report.Attendee.Email,
				report.Attendee.IDCard,
				report.Attendee.Address,
			})
		}
		firstLineID = report.OrderPayment.OrderID
	}
	return csvData
}

func GetAttendeesCSV(db *gorm.DB) [][]string {
	attendees := Attendees{}
	db.Find(&attendees)
	csvData := [][]string{{
		"ID",
		"OrderID",
		"TicketCode",
		"Sku",
		"Firstname",
		"Lastname",
		"Phone",
		"Gender",
		"Birthday",
		"Email",
		"IDCard",
		"Address",
		"ShirtSize",
	}}
	for _, attendee := range attendees {
		csvData = append(csvData, []string{
			strconv.Itoa(attendee.ID),
			strconv.Itoa(attendee.OrderID),
			attendee.TicketCode,
			attendee.Sku,
			attendee.Firstname,
			attendee.Lastname,
			attendee.Phone,
			attendee.Gender,
			attendee.Birthday,
			attendee.Email,
			attendee.IDCard,
			attendee.Address,
			attendee.ShirtSize,
		})
	}
	return csvData
}

func GetOrderItem(db *gorm.DB, orderID int) OrderItems {
	orderItems := OrderItems{}
	wpOrderItems := wp.WpWoocommerceOrderItems{}
	db.Where("order_item_type IN ('line_item','shipping') AND order_id = ?", orderID).Find(&wpOrderItems)
	for _, wpOrderItem := range wpOrderItems {
		orderItemmeta := wp.GetOrderItemmeta(db, wpOrderItem.OrderItemID)
		orderItem := &OrderItem{}
		switch wpOrderItem.OrderItemType {
		case "line_item":
			lineTotal, _ := strconv.ParseFloat(orderItemmeta["_line_total"], 64)
			orderItem.ID = wpOrderItem.OrderItemID
			orderItem.Name = wpOrderItem.OrderItemName
			orderItem.Qty = orderItemmeta["_qty"]
			orderItem.LineTotal = lineTotal

		case "shipping":
			cost, _ := strconv.ParseFloat(orderItemmeta["cost"], 64)
			orderItem.ID = wpOrderItem.OrderItemID
			orderItem.Name = wpOrderItem.OrderItemName
			orderItem.Qty = ""
			orderItem.LineTotal = cost
		}
		orderItems = append(orderItems, orderItem)
	}
	return orderItems
}

func GetShirtSizeCsv(db *gorm.DB) [][]string {
	shirtSizes := GetShirtSizeAmount(db)
	csvData := [][]string{{"ShirtSize", "Amount"}}
	for _, shirtSize := range shirtSizes {
		csvData = append(csvData, []string{
			shirtSize.ShirtSize,
			fmt.Sprintf("%d", shirtSize.Amount),
		})
	}

	csvData = append(csvData, []string{"SKU", "Stock", "Amount", "Left"})

	skus := GetSkuAmount(db)
	for _, sku := range skus {
		csvData = append(csvData, []string{
			sku.ShirtSize,
			strconv.Itoa(sku.Stock),
			fmt.Sprintf("%d", sku.Amount),
			strconv.Itoa(sku.StockLeft),
		})
	}
	return csvData
}

func IfElse(i bool, t string, f string) string {
	if i {
		return t
	} else {
		return f
	}
}
