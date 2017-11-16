package main

import (
	"strconv"

	"github.com/jinzhu/gorm"
)

func GetAttendeeReport(db *gorm.DB) [][]string {
	csvStatus := [][]string{{"Order ID", "DateTime", "Ticket Code", "Firstname", "Lastname", "Phone", "Gender", "Birthdate", "SKU"}}
	metaKeys := []string{kTicketTypeID, "ticket_code", kFirstname, kLastname, kPhone, kGender, kBirthday}
	posts := []WpPost{}
	db.Where("post_status = 'wc-processing' AND post_type = 'shop_order'").Find(&posts)

	m := NewModel(db)
	for _, post := range posts {
		var metaTickets []WpPostmeta
		db.Where("meta_key = 'ticket_code' AND meta_value LIKE ?", strconv.Itoa(post.ID)+"-%").Find(&metaTickets)
		for _, metaTicket := range metaTickets {
			postMeta := getPostMetaFields(db, metaTicket.PostID, metaKeys)
			ticketTypeID, _ := strconv.Atoi(postMeta[kTicketTypeID])
			ticketType := m.GetProduct(ticketTypeID)

			csvStatus = append(csvStatus, []string{
				strconv.Itoa(post.ID),
				post.PostDate,
				postMeta["ticket_code"],
				postMeta[kFirstname],
				postMeta[kLastname],
				postMeta[kPhone],    // Phone
				postMeta[kGender],   // Gender
				postMeta[kBirthday], // Birthdate
				ticketType.Sku,
			})
		}
	}
	return csvStatus
}

// func GetShirtSizeCsv(csvStatus [][]string) [][]string {
// 	shirtAddOnReportMap := make(map[string]int)
// 	shirtReportMap := make(map[string]int)
// 	for i, row := range csvStatus {
// 		if i > 0 {
// 			shirtAddOnReportMap[row[8]]++
// 			sku := strings.Replace(row[8], "-addon", "", -1)
// 			shirtReportMap[sku]++
// 		}
// 	}

// 	shirtReport := [][]string{}

// 	keys := make([]string, 0)
// 	for k := range shirtAddOnReportMap {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)

// 	for _, v := range keys {
// 		shirtReport = append(shirtReport, []string{v, strconv.Itoa(shirtAddOnReportMap[v])})
// 	}

// 	shirtReport = append(shirtReport, []string{"", ""})

// 	keys = make([]string, 0)
// 	for k := range shirtReportMap {
// 		keys = append(keys, k)
// 	}

// 	sort.Strings(keys)

// 	for _, v := range keys {
// 		shirtReport = append(shirtReport, []string{v, strconv.Itoa(shirtReportMap[v])})
// 	}

// 	return shirtReport
// }
