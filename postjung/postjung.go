package postjung

import (
	"log"
	"time"

	"github.com/gocolly/colly"
	"gorm.io/gorm"
)

type (
	Scraping struct {
		Collector *colly.Collector
	}
	pFunc func(_ int, elem *colly.HTMLElement)
	Forum struct {
		// FID       int    `gorm:"type:int(10);autoIncrement"`
		Fid       uint   `gorm:"primaryKey"`
		RoomId    int    `gorm:"type:Int(10)"`
		RoomName  string `gorm:"type:VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci"`
		TotalPage int    `gorm:"type:Int(10)"`
	}
	Content struct {
		Cid           uint      `gorm:"primaryKey"`
		RoomId        int       `gorm:"type:Int(10)"`
		Title         string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci"`
		Content       string    `gorm:"type:Text CHARACTER SET utf8 COLLATE utf8_general_ci"`
		CreateDate    string    `gorm:"type:Date"`
		Permalink     string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci"`
		WebsiteDomain string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci"`
		MessageType   string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci"`
		WebsiteType   string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci"`
		Author        string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci"`
		ViewCount     int       `gorm:"type:Int(10)"`
		CommentCount  int       `gorm:"type:Int(10)"`
		Tags          string    `gorm:"type:Text CHARACTER SET utf8 COLLATE utf8_general_ci"`
		PictureUrls   string    `gorm:"type:Text CHARACTER SET utf8 COLLATE utf8_general_ci"`
		ImportDate    time.Time `gorm:"type:Date"`
		UpdateDate    time.Time `gorm:"type:Date"`
	}
)

var (
	SiteConfig = map[string]string{
		"site":       "https://board.postjung.com/",
		"categories": "https://board.postjung.com/boards.php?page=",
	}
)

func (s *Scraping) Scraping(url string, findSelector string, useSelector string, p pFunc) {
	s.Collector.OnHTML(findSelector, func(e *colly.HTMLElement) {
		e.ForEach(useSelector, p)
	})
	s.Collector.Visit(url)
}

func (f *Forum) InsertToDB(db *gorm.DB) {
	forum := Forum{RoomId: f.RoomId, RoomName: f.RoomName, TotalPage: f.TotalPage}
	result := db.Create(&forum)
	log.Println("result", result)
}

func (c *Content) InsertToDB(db *gorm.DB) {
	content := Content{
		RoomId:        c.RoomId,
		Title:         c.Title,
		Content:       c.Content,
		CreateDate:    c.CreateDate,
		Permalink:     c.Permalink,
		WebsiteDomain: c.WebsiteDomain,
		MessageType:   c.MessageType,
		WebsiteType:   c.WebsiteType,
		Author:        c.Author,
		ViewCount:     c.ViewCount,
		CommentCount:  c.CommentCount,
		Tags:          c.Tags,
		PictureUrls:   c.PictureUrls,
		ImportDate:    c.ImportDate,
		UpdateDate:    c.UpdateDate,
	}
	result := db.Create(&content)
	log.Println("result", result)
}
