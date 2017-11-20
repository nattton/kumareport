package main

import (
	"github.com/code-mobi/kumareport/wp"
	"github.com/jinzhu/gorm"
)

type Model struct {
	db            *gorm.DB
	ticketTypes   map[int]Product
	skuList       map[string]Product
	attendees     map[int]Attendee
	orderPayments map[int]OrderPayment
}

type Product struct {
	ID           int `gorm:"primary_key"`
	Sku          string
	Name         string
	Stock        int
	DisplayOrder int
}

func RetrieveProducts(db *gorm.DB) {
	db.AutoMigrate(&Product{})
	model := NewModel(db)
	var posts []wp.WpPost
	db.Where("post_type = 'product_variation' AND post_status= 'publish'").Order("menu_order").Find(&posts)
	for _, post := range posts {
		p := model.GetProduct(post.ID)
		product := Product{
			ID:           post.ID,
			Sku:          p.Sku,
			Name:         post.PostTitle,
			DisplayOrder: post.MenuOrder,
		}
		db.Assign(product).FirstOrCreate(&product)
	}
}

func NewModel(db *gorm.DB) *Model {
	model := new(Model)
	model.db = db
	model.ticketTypes = make(map[int]Product)
	model.skuList = make(map[string]Product)
	model.attendees = make(map[int]Attendee)
	model.orderPayments = make(map[int]OrderPayment)
	return model
}

func (model *Model) RetieveTickets() {
	var products []Product
	model.db.Find(&products)
	for _, product := range products {
		model.ticketTypes[product.ID] = product
	}
}

func (model *Model) GetProduct(id int) Product {
	if product, ok := model.ticketTypes[id]; ok {
		return product
	}

	postMetaTicketType := wp.GetPostMetaFields(model.db, id, []string{"_sku"})
	return Product{ID: id, Sku: postMetaTicketType["_sku"]}
}

func (model *Model) RetieveAttendee() {
	var attendees []Attendee
	model.db.Find(&attendees)
	for _, attendee := range attendees {
		model.attendees[attendee.ID] = attendee
	}
}

func (model *Model) GetAttendee(id int) Attendee {
	attendee, ok := model.attendees[id]
	if ok {
		return attendee
	}

	model.db.First(&attendee, id)
	model.attendees[id] = attendee
	return attendee
}

// func (model *Model) RetieveSkuList() {
// 	var products []Product
// 	model.db.Find(&products)
// 	model.db.Where("post_type = 'product_variation' AND post_status= 'publish'").Order("menu_order").Find(&posts)
// 	for _, post := range posts {
// 		product := model.GetProduct(post.ID)
// 		model.skuList[product.Sku] = product
// 	}
// }

func (model *Model) RetieveSkuList() {
	var products []Product
	model.db.Find(&products)
	for _, product := range products {
		model.skuList[product.Sku] = product
	}
}

func (model *Model) GetSkuList(id string) Product {
	if product, ok := model.skuList[id]; ok {
		return product
	}
	return Product{}
}
