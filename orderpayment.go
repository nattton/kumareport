package main

import (
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"
)

type OrderPayment struct {
	OrderID            int `gorm:"primary_key"`
	RefID              string
	PaymentType        string
	PaymentAmount      float64
	PaymentDateTime    string
	Firstname          string
	Lastname           string
	Phone              string
	Email              string
	OrderTotal         float64
	CreditCardFee      float64
	VatFee             float64
	TaxFee             float64
	OrderCreditCardFee float64
	OrderVatFee        float64
	OrderTaxFee        float64
	OrderAfterVat      float64
	OrderAfterFee      float64
	ShippingCompany    string //_shipping_company
	ShippingAddress    string //_shipping_address_1,_shipping_address_2,_shipping_city,_shipping_state,_shipping_postcode,_shipping_country
	ShippingPostcode   string //_shipping_postcode
}

func getPostMetaOrderPayment(db *gorm.DB, postID int) OrderPayment {
	postMeta := make(map[string]string)
	postMetas := []WpPostmeta{}
	metaKeys := []string{"_order_key", "_shipping_first_name", "_shipping_last_name", "_billing_phone", "_billing_email", "_shipping_company", "_shipping_address_1", "_shipping_address_2", "_shipping_city", "_shipping_state", "_shipping_postcode", "_shipping_country", "_order_total"}
	db.Where("post_id = ? AND meta_key IN (?)", postID, metaKeys).Find(&postMetas)

	for i := range postMetas {
		postMeta[postMetas[i].MetaKey] = postMetas[i].MetaValue
	}

	refID := postMeta["_order_key"]
	paymentStatus := GetPaymentStatus(refID)
	var paymentType, paymentDateTime string
	var paymentAmount float64
	if paymentStatus.TotalRow == 1 {
		statusRow := paymentStatus.DataRow[0]
		paymentType = statusRow.PaymentType
		paymentAmount, _ = strconv.ParseFloat(statusRow.PaymentAmount, 64)
		paymentDateTime = statusRow.PaymentDateTime
	}
	orderTotal, _ := strconv.ParseFloat(postMeta["_order_total"], 64)

	order := OrderPayment{
		OrderID:          postID,
		RefID:            refID,
		Firstname:        postMeta["_shipping_first_name"],
		Lastname:         postMeta["_shipping_last_name"],
		Phone:            postMeta["_billing_phone"],
		Email:            postMeta["_billing_email"],
		OrderTotal:       orderTotal,
		PaymentType:      paymentType,
		PaymentDateTime:  paymentDateTime,
		PaymentAmount:    paymentAmount,
		ShippingCompany:  postMeta["_shipping_company"],
		ShippingPostcode: postMeta["_shipping_postcode"],
	}
	order.ShippingAddress = fmt.Sprintf("%s %s %s %s %s %s", postMeta["_shipping_address_1"], postMeta["_shipping_address_2"], postMeta["_shipping_city"], GetThailandState()[postMeta["_shipping_state"]], postMeta["_shipping_postcode"], postMeta["_shipping_country"])

	order.CalulateOrderPayment(paymentType)
	return order
}

func (order *OrderPayment) CalulateOrderPayment(paymentType string) *OrderPayment {
	order.PaymentType = paymentType
	if paymentType == "CreditCard" {
		order.CreditCardFee = 3.75
		order.VatFee = 7
		order.TaxFee = 3
		order.OrderCreditCardFee = order.OrderTotal * (order.CreditCardFee / 100)
		order.OrderVatFee = order.OrderCreditCardFee * (order.VatFee / 100)
		order.OrderTaxFee = order.OrderCreditCardFee * (order.TaxFee / 100)
		order.OrderAfterVat = order.OrderCreditCardFee + (order.OrderVatFee - order.OrderTaxFee)
		order.OrderAfterFee = order.OrderTotal - order.OrderAfterVat
	} else {
		order.OrderAfterFee = order.OrderTotal
	}
	return order
}

func NewOrderPayment(orderTotal float64, paymentType string) *OrderPayment {
	order := new(OrderPayment)
	order.OrderTotal = orderTotal
	order.PaymentType = paymentType
	if paymentType == "CreditCard" {
		order.CreditCardFee = 3.75
		order.VatFee = 7
		order.TaxFee = 3
		order.OrderCreditCardFee = orderTotal * (order.CreditCardFee / 100)
		order.OrderVatFee = order.OrderCreditCardFee * (order.VatFee / 100)
		order.OrderTaxFee = order.OrderCreditCardFee * (order.TaxFee / 100)
		order.OrderAfterVat = order.OrderCreditCardFee + (order.OrderVatFee - order.OrderTaxFee)
		order.OrderAfterFee = orderTotal - order.OrderAfterVat
	} else {
		order.OrderAfterFee = orderTotal
	}
	return order
}

func (model *Model) RetieveOrderPayment() {
	var orderPayments []OrderPayment
	model.db.Find(&orderPayments)
	for _, orderPayment := range orderPayments {
		model.orderPayments[orderPayment.OrderID] = orderPayment
	}
}

func (model *Model) GetOrderPayment(id int) OrderPayment {
	orderPayment, ok := model.orderPayments[id]
	if ok {
		return orderPayment
	}

	model.db.First(&orderPayment, id)
	model.orderPayments[id] = orderPayment
	return orderPayment
}
