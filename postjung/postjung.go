// Copyright 2021 aofiee
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package postjung

import (
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type (
	Scraping struct {
		Collector *colly.Collector
	}
	pFunc func(_ int, elem *colly.HTMLElement)
	Forum struct {
		Fid       int    `gorm:"primaryKey"`
		RoomId    int    `gorm:"type:Int(10)"`
		RoomName  string `gorm:"type:VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		TotalPage int    `gorm:"type:Int(10)"`
	}
	Content struct {
		Cid           int       `gorm:"primaryKey"`
		RoomId        int       `gorm:"type:Int(10)"`
		Title         string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		Content       string    `gorm:"type:Text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		CreateDate    string    `gorm:"type:Date"`
		Permalink     string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		WebsiteDomain string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		MessageType   string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		WebsiteType   string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		Author        string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		ViewCount     int       `gorm:"type:Int(10)"`
		CommentCount  int       `gorm:"type:Int(10)"`
		Tags          string    `gorm:"type:Text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		PictureUrls   string    `gorm:"type:Text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		ImportDate    time.Time `gorm:"type:Date"`
		UpdateDate    time.Time `gorm:"type:Date"`
	}
	Comment struct {
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
		Content       string    `gorm:"type:Text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		CommentDate   string    `gorm:"type:VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		Permalink     string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		WebsiteDomain string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		CommentType   string    `gorm:"type:VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		Author        string    `gorm:"type:VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		ContentId     int       `gorm:"type:Int(10)"`
		ViewCount     int       `gorm:"type:Int(10)"`
		CommentCount  int       `gorm:"type:Int(10)"`
		Tags          string    `gorm:"type:Text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
		PictureUrls   string    `gorm:"type:Text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
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

func (s *Scraping) ScrapingCount(url string, findSelector string, useSelector string) string {
	var total string
	s.Collector.OnHTML(findSelector, func(e *colly.HTMLElement) {
		var hrefText []string
		e.ForEach(useSelector, func(_ int, elem *colly.HTMLElement) {
			hrefText = append(hrefText, elem.Text)
		})
		total = strings.Join(hrefText[len(hrefText)-2:len(hrefText)-1], "")
	})
	s.Collector.Visit(url)
	return total
}
