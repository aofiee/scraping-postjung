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

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aofiee/scraping/postjung"
	"github.com/fatih/color"
	"github.com/gocolly/colly"

	"github.com/PuerkitoBio/goquery"
	emoji "github.com/tmdvs/Go-Emoji-Utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	chBuf = 100
)

type (
	scrapTopic struct {
		text    string
		link    string
		comment string
	}
	scrapComment struct {
		text      string
		author    string
		commentNo string
	}
)

var (
	site = map[string]string{
		"postjung": "https://board.postjung.com/",
	}
	categoriesURL string
	red           = color.New(color.FgRed, color.Bold).SprintFunc()
	green         = color.New(color.FgGreen, color.Bold).SprintFunc()
	blue          = color.New(color.FgBlue, color.Bold).SprintFunc()
	cyan          = color.New(color.FgCyan, color.Bold).SprintFunc()
	totalURL      []string
	db            *gorm.DB
	dbError       error
)

func main() {
	log.SetFlags(log.Ltime)
	if arg() {
		dsn := "root:helloworld@tcp(127.0.0.1:3306)/postjung?charset=utf8mb4&parseTime=True&loc=Local"
		db, dbError = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if dbError != nil {
			log.Fatal(dbError)
		}
		initWithCmd()
	}

}

func initWithCmd() {
	switch os.Args[1] {
	case "init":
		initForums()
		break
	case "list-all-room":
		listAllRooms()
		break
	case "get-topic":
		roomID := os.Args[2]
		getAllTopicFrom(roomID)
		break
	case "get-comment":
		roomID := os.Args[2]
		getAllCommentsFrom(roomID)
		break
	}
}

func initForums() {
	categoriesURL = postjung.SiteConfig["categories"]
	i := 0
	var chk bool
	chk = true
	for chk {
		loadInitPage(categoriesURL+strconv.Itoa(i), &chk)
		i++
	}
}

func listAllRooms() {
	var forum []postjung.Forum
	if err := db.Find(&forum).Error; err != nil {
		log.Fatalln(err)
	}
	for _, r := range forum {
		log.Println(green("Room ID: "), cyan(strconv.Itoa(r.RoomId))+" "+red(r.RoomName))
	}
}

func getAllTopicFrom(roomID string) {
	var forum postjung.Forum
	if err := db.Where("room_id = ?", roomID).First(&forum).Error; err != nil {
		log.Fatalln(err)
	}
	for page := 0; page < forum.TotalPage; page++ {
		chGettopicLinkFromPage := make(chan scrapTopic, chBuf)
		chGettopicLinkFromPageIsDone := make(chan bool, chBuf)
		go findAllTopic("https://board.postjung.com/board.php?id="+roomID+"&page="+strconv.Itoa(page), chGettopicLinkFromPage)
		go getContentsFromTopic(chGettopicLinkFromPage, chGettopicLinkFromPageIsDone, roomID)
		log.Println(green("Done"), <-chGettopicLinkFromPageIsDone)
	}
}

func getAllCommentsFrom(roomID string) {
	var content []postjung.Content
	if err := db.Find(&content).Where("comment_count > 0 AND room_id = ?", roomID).Error; err != nil {
		log.Fatalln(err)
	}
	for _, r := range content {
		chGetComment := make(chan scrapComment, chBuf)
		log.Println(cyan("content "), r.Permalink)
		go findAllCommentFromLink(r.Cid, r.Permalink, chGetComment)
		log.Println(cyan("Comment"), <-chGetComment)
	}
}

func findAllCommentFromLink(contentId int, link string, ch chan<- scrapComment) {
	findComment := postjung.Scraping{colly.NewCollector()}
	findComment.Scraping(link, "body > div.mainbox", "script:nth-child(9)", func(_ int, elem *colly.HTMLElement) {
		js := strings.ReplaceAll(strings.ReplaceAll(elem.Text, "var cmnvar=", ""), ";", "")
		ch <- scrapComment{}
		comments := postjung.Comment{}
		json.Unmarshal([]byte(js), &comments)
		param := "?cmkey=" + comments.Cmkey + "&owner=" + strconv.Itoa(comments.Owner) + "&page=0&notop=1&noadd=1&maxlist=10&adsense_allow=1&adsense_allow_ns=1"

		res, err := http.Get(postjung.SiteConfig["comments"] + param)
		if err != nil {
			log.Fatal(err)
		}
		parserHTMLCommentsFromData(contentId, link, res)

	})
	defer close(ch)
}

func parserHTMLCommentsFromData(contentId int, link string, res *http.Response) {
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("div.cm").Each(func(i int, s *goquery.Selection) {
		author := s.Find(".cm > div.xbody > a.xname").Text()
		comment := s.Find(".cm > div.xbody > div.xtext").Text()
		publistDate := s.Find(".cm > div.xbody > a.xtoolbt")
		onclickAttr, _ := publistDate.Attr("onclick")
		onclickAttr = strings.ReplaceAll(strings.ReplaceAll(onclickAttr, "cmn.tool(this,", ""), "); return false;", "")
		commentDate := postjung.CommentDate{}
		json.Unmarshal([]byte(onclickAttr), &commentDate)
		insertComment := postjung.CommentContent{
			Content:       emoji.RemoveAll(comment),
			CommentDate:   commentDate.Unixtime,
			Permalink:     link,
			WebsiteDomain: postjung.SiteConfig["site"],
			CommentType:   "comment",
			Author:        author,
			ContentId:     contentId,
			ViewCount:     0,
			CommentCount:  0,
			Tags:          "",
			PictureUrls:   "",
			CreateDate:    time.Now(),
			UpdateDate:    time.Now(),
		}
		result := db.Create(&insertComment)
		log.Println("result", result)
		chReplyComments := make(chan string, chBuf)
		go parserHTMLReplyCommentsFromData(chReplyComments, s, contentId, link)
		log.Println("reply comment : ", <-chReplyComments)
	})
	defer func() {
		res.Body.Close()
	}()
}

func parserHTMLReplyCommentsFromData(ch chan<- string, s *goquery.Selection, contentId int, link string) {
	s.Find(".cm > div.xbody > div.reps > div.rep").Each(func(j int, e *goquery.Selection) {
		ed := e.Find("div.rep > a.xtoolbt")
		edAttr, _ := ed.Attr("onclick")
		repsContent := e.Find("div.rep > div.reptext").Text()
		ch <- repsContent
		edAttr = strings.ReplaceAll(strings.ReplaceAll(edAttr, "cmn.tool(this,", ""), "); return false;", "")
		repsDate := postjung.CommentDate{}
		json.Unmarshal([]byte(edAttr), &repsDate)
		insertReps := postjung.CommentContent{
			Content:       emoji.RemoveAll(repsContent),
			CommentDate:   repsDate.Unixtime,
			Permalink:     link,
			WebsiteDomain: postjung.SiteConfig["site"],
			CommentType:   "reps",
			Author:        strconv.Itoa(repsDate.Userid),
			ContentId:     contentId,
			ViewCount:     0,
			CommentCount:  0,
			Tags:          "",
			PictureUrls:   "",
			CreateDate:    time.Now(),
			UpdateDate:    time.Now(),
		}
		result := db.Create(&insertReps)
		log.Println("result", result)

	})
	defer close(ch)
}

func getContentsFromTopic(ch <-chan scrapTopic, done chan bool, roomID string) {
	for c := range ch {
		c := c
		getContent := postjung.Scraping{colly.NewCollector()}
		getContent.Scraping(c.link, "body", ".mainbox", func(_ int, elem *colly.HTMLElement) {
			title := emoji.RemoveAll(elem.ChildText("body > div.mainbox > h1"))
			author := emoji.RemoveAll(elem.ChildText("#maincontent > #hbar1 > a:nth-child(2)"))
			allContent := emoji.RemoveAll(strings.ReplaceAll(elem.ChildText("#maincontent"), "'", ""))
			imgs := elem.ChildAttrs("img", "src")
			imgsJSON, _ := json.MarshalIndent(imgs, "", "    ")
			tags := []string{}
			elem.ForEach("#maincontent > div.sptags > a", func(_ int, t *colly.HTMLElement) {
				tags = append(tags, emoji.RemoveAll(t.Text))
			})
			tagsStore := strings.Join(tags[:], ",")
			publishDate := elem.ChildAttr("#infobox > div.spinfo.spinfo1 > div.xbody > div > time:nth-child(1)", "datetime")
			comments := strings.ReplaceAll(
				strings.Join(
					strings.Split(
						strings.ReplaceAll(
							strings.ReplaceAll(c.comment, "(", ""), ")", ""), ",")[:1],
					""), "ตอบ", "")
			if comments == "" {
				comments = "0"
			}
			messageType := "post"
			websiteType := "news"
			RID, _ := strconv.Atoi(roomID)
			commentCount, err := strconv.Atoi(strings.TrimSpace(comments))
			if err != nil {
				log.Println(red("error "), err)
			}
			content := postjung.Content{
				RoomId:        RID,
				Title:         title,
				Content:       allContent,
				CreateDate:    publishDate,
				Permalink:     c.link,
				WebsiteDomain: postjung.SiteConfig["site"],
				MessageType:   messageType,
				WebsiteType:   websiteType,
				Author:        author,
				ViewCount:     0,
				CommentCount:  commentCount,
				Tags:          tagsStore,
				PictureUrls:   string(imgsJSON),
				ImportDate:    time.Now(),
				UpdateDate:    time.Now(),
			}
			result := db.Create(&content)
			log.Println("result", result)
		})

	}
	defer func() {
		done <- true
		close(done)
	}()
}

func loadInitPage(categoriesURL string, chk *bool) {
	forumPage := postjung.Scraping{colly.NewCollector()}
	forumPage.Scraping(categoriesURL, "div.pagebar", "a.xnav", func(_ int, elem *colly.HTMLElement) {
		if elem.Text == "next >" {
			findForumRoom(categoriesURL)
			*chk = true
		} else {
			*chk = false
		}
	})

}

func findForumRoom(link string) {
	forumRoom := postjung.Scraping{colly.NewCollector()}
	forumRoom.Scraping(link, "body > div.splist", "a", func(_ int, elem *colly.HTMLElement) {
		link := elem.Attr("href")
		roomID := strings.Join(strings.Split(link, "-")[1:2], "")
		total := findAllTopicPage(postjung.SiteConfig["site"] + "board.php?id=" + roomID)
		totalPage, _ := strconv.Atoi(total)
		RID, err := strconv.Atoi(roomID)
		if err != nil {
			log.Fatal(red("error :"), err)
		}
		forum := postjung.Forum{
			RoomName:  emoji.RemoveAll(elem.Text),
			TotalPage: totalPage,
			RoomId:    RID,
		}
		result := db.Create(&forum)
		log.Println("result", result)

	})
}

func findAllTopicPage(link string) string {
	topicPage := postjung.Scraping{colly.NewCollector()}
	total := topicPage.ScrapingCount(link, "div.pagebar", "a")
	return total
}

func findAllTopic(link string, ch chan<- scrapTopic) {
	topic := postjung.Scraping{colly.NewCollector()}
	page := link[len(link)-1:]
	var selectorTarget string
	selectorTarget = "body > div.mainbox > div.splist"
	if page == "0" {
		selectorTarget = "body > div.mainbox > div.sphot > div.splist"
	}
	topic.Scraping(link, selectorTarget, "a", func(_ int, elem *colly.HTMLElement) {
		l := elem.Attr("href")
		comment := elem.ChildText("span > span.xinfo")
		pack := scrapTopic{
			text:    elem.Text,
			link:    postjung.SiteConfig["site"] + l,
			comment: comment,
		}
		ch <- pack
		log.Println(green(elem.Text) + " " + l)
	})
	defer close(ch)
}
func arg() bool {
	if len(os.Args) < 2 {
		log.Println(red("Missing argument"))
		os.Exit(1)
	}
	return true
}
