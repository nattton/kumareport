package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

const (
	kTicketCode   = "ticket_code"
	kFirstname    = "first_name"
	kLastname     = "last_name"
	kTicketTypeID = "ticket_type_id"
	kPhone        = "_tcfn_42"
	kGender       = "_tcfn_5920"
	kBirthday     = "_tcfn_3326"
	kEmail        = "email_tcfn_4728"
	kIDCard       = "_tcfn_2214"
	kAddress      = "_tcfn_207"
)

type Attendee struct {
	ID           int `gorm:"primary_key"`
	OrderID      int
	TicketTypeID int //ticket_type_id
	TicketCode   string
	Firstname    string //first_name
	Lastname     string //last_name
	Phone        string //_tcfn_42
	Gender       string //_tcfn_5920
	Birthday     string //_tcfn_3326
	Email        string //email_tcfn_4728
	IDCard       string //_tcfn_2214
	Address      string //_tcfn_207
	Sku          string
	ShirtSize    string
}

func AttendeeHandler(c *gin.Context) {
	id := c.Param("id")
	attendeeID, err := strconv.Atoi(id)
	if err != nil {
		NotFoundHandler(c, err.Error())
	}
	db, _ := OpenDB()
	defer db.Close()

	m := NewModel(db)
	m.RetieveTickets()
	attendee := GetAttendee(db, m, attendeeID)
	c.HTML(http.StatusOK, "attendee.tmpl", gin.H{
		"attendee":    attendee,
		"ticketTypes": m.ticketTypes,
	})
}

func AttendeeUpdateHandler(c *gin.Context) {
	id := c.Param("id")
	var formA Attendee
	if err := c.ShouldBind(&formA); err != nil {
		NotFoundHandler(c, err.Error())
		return
	}
	if id != strconv.Itoa(formA.ID) {
		NotFoundHandler(c, "Bad Request")
		return
	}

	log.Printf("%v", formA)

	db, _ := OpenDB()
	m := NewModel(db)
	m.RetieveTickets()
	attendee := GetAttendee(db, m, formA.ID)
	if formA.ID != attendee.ID {
		NotFoundHandler(c, "Not Found")
		return
	}

	if formA.TicketTypeID != attendee.TicketTypeID {
		db.Exec("UPDATE wp_postmeta SET meta_value= ? WHERE meta_key=? AND post_id = ?", formA.TicketTypeID, kTicketTypeID, attendee.ID)
	}
	if formA.Firstname != attendee.Firstname {
		db.Exec("UPDATE wp_postmeta SET meta_value= ? WHERE meta_key=? AND post_id = ?", formA.Firstname, kFirstname, attendee.ID)
	}
	if formA.Lastname != attendee.Lastname {
		db.Exec("UPDATE wp_postmeta SET meta_value= ? WHERE meta_key=? AND post_id = ?", formA.Lastname, kLastname, attendee.ID)
	}
	if formA.Lastname != attendee.Lastname {
		db.Exec("UPDATE wp_postmeta SET meta_value= ? WHERE meta_key=? AND post_id = ?", formA.Lastname, kLastname, attendee.ID)
	}
	if formA.Phone != attendee.Phone {
		db.Where(WpPostmeta{PostID: attendee.ID, MetaKey: kPhone}).Assign(WpPostmeta{MetaValue: formA.Phone}).FirstOrCreate(&WpPostmeta{})
	}
	if formA.Gender != attendee.Gender {
		db.Where(WpPostmeta{PostID: attendee.ID, MetaKey: kGender}).Assign(WpPostmeta{MetaValue: formA.Gender}).FirstOrCreate(&WpPostmeta{})
	}
	if formA.Birthday != attendee.Birthday {
		db.Where(WpPostmeta{PostID: attendee.ID, MetaKey: kBirthday}).Assign(WpPostmeta{MetaValue: formA.Birthday}).FirstOrCreate(&WpPostmeta{})
	}
	if formA.Address != attendee.Address {
		db.Where(WpPostmeta{PostID: attendee.ID, MetaKey: kAddress}).Assign(WpPostmeta{MetaValue: formA.Address}).FirstOrCreate(&WpPostmeta{})
	}

	if formA.Email != attendee.Email && formA.Email != "" {
		db.Where(WpPostmeta{PostID: attendee.ID, MetaKey: kEmail}).Assign(WpPostmeta{MetaValue: formA.Email}).FirstOrCreate(&WpPostmeta{})
	}
	if formA.Email == "" {
		UpdatePostmetaEmpty(db, attendee.ID, kEmail)
	}

	if formA.IDCard != attendee.IDCard && formA.IDCard != "" {
		db.Where(WpPostmeta{PostID: attendee.ID, MetaKey: kIDCard}).Assign(WpPostmeta{MetaValue: formA.IDCard}).FirstOrCreate(&WpPostmeta{})
	}
	if formA.IDCard == "" {
		UpdatePostmetaEmpty(db, attendee.ID, kIDCard)
	}

	attendee = GetAttendee(db, m, formA.ID)
	db.Save(&attendee)

	c.HTML(http.StatusOK, "attendee.tmpl", gin.H{
		"attendee":    attendee,
		"ticketTypes": m.ticketTypes,
		"message":     "Save Complete",
	})
}

func UpdatePostmetaEmpty(db *gorm.DB, attendeeID int, metaKey string) {
	db.Exec("UPDATE wp_postmeta SET meta_value= ? WHERE meta_key=? AND post_id = ?", "", metaKey, attendeeID)
}

func GenerateAttendee(db *gorm.DB, reUpdate bool) {
	db.AutoMigrate(&Attendee{})

	m := NewModel(db)
	m.RetieveTickets()
	m.RetieveAttendee()
	posts := []WpPost{}
	if reUpdate {
		db.Where("post_status = 'wc-processing' AND post_type = 'shop_order'").Find(&posts)
	} else {
		db.Raw("SELECT wp_posts.* FROM wp_posts LEFT OUTER JOIN attendees ON (wp_posts.ID = attendees.order_id) WHERE wp_posts.post_status = 'wc-processing' AND wp_posts.post_type = 'shop_order' AND attendees.id IS NULL").Scan(&posts)
	}
	for _, post := range posts {
		var metaTickets []WpPostmeta
		db.Where("meta_key = ? AND meta_value LIKE ?", kTicketCode, strconv.Itoa(post.ID)+"-%").Find(&metaTickets)
		var wg sync.WaitGroup
		for _, metaTicket := range metaTickets {
			throttle <- 1
			wg.Add(1)
			go UpdateAttendee(db, &wg, throttle, m, post, metaTicket.PostID, reUpdate)
		}
		wg.Wait()
	}
}

func UpdateAttendee(db *gorm.DB, wg *sync.WaitGroup, throttle chan int, m *Model, post WpPost, attendeeID int, reUpdate bool) {
	defer wg.Done()
	var isNew bool
	attendee := m.GetAttendee(attendeeID)
	if attendee.ID != attendeeID {
		isNew = true
		attendee.ID = attendeeID
	} else {
		if !reUpdate {
			return
		}
	}
	attendee = GetAttendee(db, m, attendeeID)
	attendee.OrderID = post.ID
	if isNew {
		db.Create(&attendee)
	} else {
		db.Save(&attendee)
	}
	<-throttle
}

func GetAttendees(db *gorm.DB, orderID int) []Attendee {
	attendees := []Attendee{}
	m := NewModel(db)
	m.RetieveAttendee()
	var metaTickets []WpPostmeta
	db.Where("meta_key = 'ticket_code' AND meta_value LIKE ?", strconv.Itoa(orderID)+"-%").Find(&metaTickets)
	for _, metaTicket := range metaTickets {
		attendees = append(attendees, GetAttendee(db, m, metaTicket.PostID))
	}
	return attendees
}

func GetAttendee(db *gorm.DB, m *Model, attendeeID int) Attendee {
	var attendee Attendee
	db.First(&attendee, attendeeID)
	if attendee.ID != attendeeID {
		attendee.ID = attendeeID
	}

	attendee = getPostMetaAttendee(db, attendeeID)
	ticketType := m.GetProduct(attendee.TicketTypeID)
	attendee.Sku = ticketType.Sku
	orderID, _ := strconv.Atoi(strings.Split(attendee.TicketCode, "-")[0])
	attendee.OrderID = orderID
	attendee.ShirtSize = strings.Replace(ticketType.Sku, "-addon", "", -1)
	return attendee
}

type ShirtSizeAmount struct {
	ShirtSize string
	Amount    int
	Stock     int
	StockLeft int
}

func GetShirtSizeAmount(db *gorm.DB) []ShirtSizeAmount {
	rows, _ := db.Raw("SELECT shirt_size, COUNT(*) FROM attendees GROUP BY shirt_size ORDER BY shirt_size").Rows()
	defer rows.Close()
	shirtSizes := []ShirtSizeAmount{}
	for rows.Next() {
		shirtSizeAmount := ShirtSizeAmount{}
		rows.Scan(&shirtSizeAmount.ShirtSize, &shirtSizeAmount.Amount)
		shirtSizes = append(shirtSizes, shirtSizeAmount)
	}
	return shirtSizes
}

func GetSkuAmount(db *gorm.DB) []ShirtSizeAmount {
	rows, _ := db.Raw("SELECT attendees.sku, products.stock, COUNT(*), (products.stock-COUNT(*)) stock_left FROM attendees INNER JOIN products ON (attendees.sku = products.sku) WHERE products.sku LIKE 'kuma%' GROUP BY `sku` ORDER BY display_order").Rows()
	defer rows.Close()
	skus := []ShirtSizeAmount{}
	for rows.Next() {
		shirtSizeAmount := ShirtSizeAmount{}
		rows.Scan(&shirtSizeAmount.ShirtSize, &shirtSizeAmount.Stock, &shirtSizeAmount.Amount, &shirtSizeAmount.StockLeft)
		skus = append(skus, shirtSizeAmount)
	}
	rows, _ = db.Raw("SELECT attendees.sku, products.stock, COUNT(*), (products.stock-COUNT(*)) stock_left FROM attendees INNER JOIN products ON (attendees.sku = products.sku) WHERE products.sku LIKE 'sister%' GROUP BY `sku` ORDER BY display_order").Rows()
	defer rows.Close()
	for rows.Next() {
		shirtSizeAmount := ShirtSizeAmount{}
		rows.Scan(&shirtSizeAmount.ShirtSize, &shirtSizeAmount.Stock, &shirtSizeAmount.Amount, &shirtSizeAmount.StockLeft)
		skus = append(skus, shirtSizeAmount)
	}
	return skus
}
