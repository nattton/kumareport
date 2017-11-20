package wp

import (
	"github.com/jinzhu/gorm"
)

type WpUser struct {
	ID           int `gorm:"primary_key;column:ID"`
	UserLogin    string
	UserPass     string
	UserNicename string
	UserEmail    string
	DisplayName  string
}

type WpPost struct {
	ID         int `gorm:"primary_key;column:ID"`
	PostDate   string
	PostTitle  string
	PostStatus string
	PostType   string
	MenuOrder  int
}

type WpPostmeta struct {
	MetaID    int `gorm:"primary_key"`
	PostID    int
	MetaKey   string
	MetaValue string
}

type WpWoocommerceOrderItem struct {
	OrderItemID   int `gorm:"primary_key"`
	OrderItemName string
	OrderItemType string
	OrderID       int
}

type WpWoocommerceOrderItemmeta struct {
	MetaID      int `gorm:"primary_key"`
	OrderItemID int
	MetaKey     string
	MetaValue   string
}

func GetPostMetaFields(db *gorm.DB, postID int, metaKeys []string) map[string]string {
	postMeta := make(map[string]string)
	postMetas := []WpPostmeta{}
	db.Where("post_id = ? AND meta_key IN (?)", postID, metaKeys).Find(&postMetas)

	for i := range postMetas {
		postMeta[postMetas[i].MetaKey] = postMetas[i].MetaValue
	}
	return postMeta
}

func GetOrderItemmeta(db *gorm.DB, orderItemID int) map[string]string {
	postMeta := make(map[string]string)
	postMetas := []WpWoocommerceOrderItemmeta{}
	metaKeys := []string{"_qty", "_line_total"}
	db.Where("order_item_id = ? AND meta_key IN (?)", orderItemID, metaKeys).Find(&postMetas)

	for i := range postMetas {
		postMeta[postMetas[i].MetaKey] = postMetas[i].MetaValue
	}
	return postMeta
}
