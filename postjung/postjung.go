package postjung

import (
	"time"

	"github.com/gocolly/colly"
)

type (
	Scraping struct {
		Collector *colly.Collector
	}
	pFunc func(_ int, elem *colly.HTMLElement)
	Forum struct {
		// FID       int    `gorm:"type:int(10);autoIncrement"`
		Fid       int    `gorm:"primaryKey"`
		RoomId    int    `gorm:"type:Int(10)"`
		RoomName  string `gorm:"type:VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci"`
		TotalPage int    `gorm:"type:Int(10)"`
	}
	Content struct {
		Cid           int       `gorm:"primaryKey"`
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
	Comment struct {
		/*
			{"cmkey":"webboard:1225437","owner":1543284,"inittime":1615446708,"userid":0,"user":"","delable":false,"hidable":false,"custom_title":"","cust_middle":"<a href=\"https:\/\/board.postjung.com\/1225437\" class=\"sptitle2\">\u0e04\u0e23\u0e39\u0e2a\u0e2d\u0e19\u0e19\u0e31\u0e01\u0e40\u0e23\u0e35\u0e22\u0e19 !! \u0e42\u0e25\u0e01\u0e19\u0e35\u0e49\u0e15\u0e31\u0e14\u0e2a\u0e34\u0e19\u0e04\u0e19\u0e08\u0e32\u0e01\u0e01\u0e32\u0e23\u0e17\u0e33\u0e1c\u0e34\u0e14\u0e40\u0e1e\u0e35\u0e22\u0e07\u0e04\u0e23\u0e31\u0e49\u0e07\u0e40\u0e14\u0e35\u0e22\u0e27!!<\/a>","useronly":true}
		*/
		Cmkey       string `json:"cmkey"`
		Owner       int    `json:"owner"`
		Inittime    int    `json:"inittime"`
		Userid      int    `json:"userid"`
		User        string `json:"user"`
		Delable     bool   `json:"delable"`
		Hidable     bool   `json:"hidable"`
		CustomTitle string `json:"custom_title"`
		CustMiddle  string `json:"cust_middle"`
		Useronly    bool   `json:"useronly"`
	}
	CommentDate struct {
		Cmid     string `json:cmid`
		Unixtime string `json:unixtime`
		Userid   int    `json:userid`
		Ip       string `json:ip`
	}
	CommentContent struct {
		Cid           uint      `gorm:"primaryKey"`
		Content       string    `gorm:"type:Text CHARACTER SET utf8 COLLATE utf8_general_ci"`
		CommentDate   string    `gorm:"type:VARCHAR(50) CHARACTER SET utf8 COLLATE utf8_general_ci"`
		Permalink     string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci"`
		WebsiteDomain string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci"`
		CommentType   string    `gorm:"type:VARCHAR(20) CHARACTER SET utf8 COLLATE utf8_general_ci"`
		Author        string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci"`
		ContentId     int       `gorm:"type:Int(10)"`
		ViewCount     int       `gorm:"type:Int(10)"`
		CommentCount  int       `gorm:"type:Int(10)"`
		Tags          string    `gorm:"type:Text CHARACTER SET utf8 COLLATE utf8_general_ci"`
		PictureUrls   string    `gorm:"type:Text CHARACTER SET utf8 COLLATE utf8_general_ci"`
		CreateDate    time.Time `gorm:"type:Date"`
		UpdateDate    time.Time `gorm:"type:Date"`
	}
)

var (
	SiteConfig = map[string]string{
		"site":       "https://board.postjung.com/",
		"categories": "https://board.postjung.com/boards.php?page=",
		"comments":   "https://board.postjung.com/wwwroot/cmn/inc.loadcm.ajax.php",
	}
)

func (s *Scraping) Scraping(url string, findSelector string, useSelector string, p pFunc) {
	s.Collector.OnHTML(findSelector, func(e *colly.HTMLElement) {
		e.ForEach(useSelector, p)
	})
	s.Collector.Visit(url)
}
