package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
)

func WriteCSVFile(csvRow [][]string, fileName string) {
	csvfile, err := os.Create(fileName + ".csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer csvfile.Close()

	w := csv.NewWriter(csvfile)
	w.WriteAll(csvRow) // calls Flush internally

	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
}

func GenerateOrderCSVFile(db *gorm.DB) {
	WriteCSVFile(GetOrderReport(db), GetFileNameNow("order"))
}

func GenerateShirtSizeCSVFile(db *gorm.DB) {
	WriteCSVFile(GetShirtSizeCsv(db), GetFileNameNow("shirt_size"))
}

func GenerateReportOrderPayment(db *gorm.DB) {
	db.AutoMigrate(&OrderPayment{})

	posts := []WpPost{}
	db.Where("post_status = 'wc-processing' AND post_type = 'shop_order'").Find(&posts)
	m := NewModel(db)
	if len(posts) > 0 {
		m.RetieveOrderPayment()
	}
	for _, post := range posts {
		orderPayment := m.GetOrderPayment(post.ID)
		if orderPayment.OrderID != post.ID {
			orderPayment := getPostMetaOrderPayment(db, post.ID)
			db.Create(&orderPayment)
		}
	}
}

func GetReportOrderPayment(db *gorm.DB) [][]string {
	orderPayments := []OrderPayment{}
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

func GetReportOrders(db *gorm.DB) [][]string {
	orderPayments := []OrderPayment{}
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
	// for _, order := range orderPayments {
	// 	var attendees []Attendee
	// 	db.Where("order_id = ?", order.OrderID).Order("id").Find(&attendees)
	// 	for i, attendee := range attendees {
	// 		if i == 0 {
	// 			csvData = append(csvData, []string{
	// 				strconv.Itoa(order.OrderID),
	// 				order.Firstname,
	// 				order.Lastname,
	// 				order.Phone,
	// 				Float64ToString(order.OrderTotal),
	// 				order.ShippingCompany,
	// 				order.ShippingAddress,
	// 				order.ShippingPostcode,
	// 				strconv.Itoa(attendee.ID),
	// 				attendee.Firstname,
	// 				attendee.Lastname,
	// 				attendee.Phone,
	// 				attendee.Gender,
	// 				attendee.Birthday,
	// 				attendee.Email,
	// 				attendee.IDCard,
	// 				attendee.Address,
	// 				attendee.Sku,
	// 			})
	// 		} else {
	// 			csvData = append(csvData, []string{
	// 				"", "", "", "", "", "", "", "",
	// 				attendee.Firstname,
	// 				attendee.Lastname,
	// 				attendee.Phone,
	// 				attendee.Gender,
	// 				attendee.Birthday,
	// 				attendee.Email,
	// 				attendee.IDCard,
	// 				attendee.Address,
	// 				attendee.Sku,
	// 			})
	// 		}
	// 	}
	// }
	return csvData
}

func GetAttendeesCSV(db *gorm.DB) [][]string {
	attendees := []Attendee{}
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

func GetOrderReport(db *gorm.DB) [][]string {
	csvStatus := [][]string{{"Order ID", "RefID", "Payment Method", "DateTime", "Item Name", "Quantity", "Line Total", "Firstname", "Lastname", "Phone", "Shipping Address", "OrderTotal", "Payment Type", "Payment DateTime", "Payment Amount"}}

	posts := []WpPost{}
	db.Where("post_status = 'wc-processing' AND post_type = 'shop_order'").Find(&posts)

	for _, post := range posts {
		postMeta := getPostMeta(db, post.ID)
		refID := postMeta["_order_key"]
		paymentStatus := GetPaymentStatus(refID)
		var paymentType, paymentDateTime, paymentAmount string
		if paymentStatus.TotalRow == 1 {
			statusRow := paymentStatus.DataRow[0]
			paymentType = statusRow.PaymentType
			paymentDateTime = statusRow.PaymentDateTime
			paymentAmount = statusRow.PaymentAmount
		}

		orderItems := []WpWoocommerceOrderItem{}
		db.Where("order_item_type = 'line_item' AND order_id = ?", post.ID).Find(&orderItems)
		for orderItemIndex, orderItem := range orderItems {

			orderItemmeta := getOrderItemmeta(db, orderItem.OrderItemID)
			csvStatus = append(csvStatus, []string{
				strconv.Itoa(post.ID),
				IfElse(orderItemIndex == 0, refID, ""),
				IfElse(orderItemIndex == 0, postMeta["_payment_method"], ""),
				post.PostDate,
				orderItem.OrderItemName,
				orderItemmeta["_qty"],
				orderItemmeta["_line_total"],
				IfElse(orderItemIndex == 0, postMeta["_shipping_first_name"], ""),
				IfElse(orderItemIndex == 0, postMeta["_shipping_last_name"], ""),
				IfElse(orderItemIndex == 0, postMeta["_billing_phone"], ""),
				IfElse(orderItemIndex == 0, postMeta["_shipping_address_index"], ""),
				IfElse(orderItemIndex == 0, postMeta["_order_total"], ""),
				IfElse(orderItemIndex == 0, paymentType, ""),
				IfElse(orderItemIndex == 0, paymentDateTime, ""),
				IfElse(orderItemIndex == 0, paymentAmount, ""),
			})
		}
	}
	return csvStatus
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
