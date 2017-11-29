package wp

import (
	"github.com/jinzhu/gorm"
)

type WpPost struct {
	ID         int `gorm:"primary_key;column:ID"`
	PostDate   string
	PostTitle  string
	PostStatus string
	PostType   string
	MenuOrder  int
}

type WpPosts []*WpPost

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

type WpWoocommerceOrderItems []*WpWoocommerceOrderItem

type WpWoocommerceOrderItemmeta struct {
	MetaID      int `gorm:"primary_key"`
	OrderItemID int
	MetaKey     string
	MetaValue   string
}

type WpWoocommerceOrderItemmetas []*WpWoocommerceOrderItemmeta

type WpComment struct {
	CommentID      int    `gorm:"primary_key;column:comment_ID"`
	CommentPostID  string `gorm:"column:comment_post_ID"`
	CommentContent string
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
	postMetas := WpWoocommerceOrderItemmetas{}
	metaKeys := []string{"_qty", "_line_total", "cost"}
	db.Where("order_item_id = ? AND meta_key IN (?)", orderItemID, metaKeys).Find(&postMetas)

	for i := range postMetas {
		postMeta[postMetas[i].MetaKey] = postMetas[i].MetaValue
	}
	return postMeta
}
