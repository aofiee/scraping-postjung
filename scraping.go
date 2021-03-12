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
	chBuf = 5
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
	target        string
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
		dsn := "root:helloworld@tcp(127.0.0.1:3306)/postjung?charset=utf8&parseTime=True&loc=Local"
		db, dbError = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if dbError != nil {
			log.Fatal(dbError)
		}
		setTargetURL(&target)
	}

}

func setTargetURL(t *string) {
	*t = os.Args[1]
	switch os.Args[1] {
	case "init":
		log.Println(green("extract all forum from "), *t)
		categoriesURL = postjung.SiteConfig["categories"]
		i := 0
		var chk bool
		chk = true
		for chk {
			loadInitPage(categoriesURL+strconv.Itoa(i), &chk)
			i++
		}

		break
	case "list-all-room":
		var forum []postjung.Forum
		if err := db.Find(&forum).Error; err != nil {
			log.Fatalln(err)
		}
		for _, r := range forum {
			log.Println(green("Room id : "), strconv.Itoa(r.RoomId)+" "+red(r.RoomName))
		}
		break
	case "gettopic":
		roomID := os.Args[2]
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

		break
	case "get-comment":
		roomID := os.Args[2]
		room, _ := strconv.Atoi(roomID)
		var content []postjung.Content
		if err := db.Find(&content).Where("comment_count > 0 AND room_id = ?", room).Error; err != nil {
			log.Fatalln(err)
		}
		for _, r := range content {
			chGetComment := make(chan scrapComment, chBuf)
			// chGetCommentIsDone := make(chan bool, chBuf)
			log.Println(cyan("content "), r.Permalink)
			findAllCommentFromLink(r.Cid, r.Permalink, chGetComment)
			log.Println(cyan("Hello"), <-chGetComment)
		}
		break
	case "test":
		var forum postjung.Forum
		if err := db.Find(&forum).Error; err != nil {
			log.Fatalln(err)
		}
		log.Println(green(forum))
		break
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
		scrapingHTMLFromData(contentId, link, res)

	})
	defer close(ch)
}

func scrapingHTMLFromData(contentId int, link string, res *http.Response) {
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("div.cm").Each(func(i int, s *goquery.Selection) {
		log.Println(cyan("comment"), s)
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
		s.Find(".cm > div.xbody > div.reps > div.rep").Each(func(j int, e *goquery.Selection) {
			ed := e.Find("div.rep > a.xtoolbt")
			edAttr, _ := ed.Attr("onclick")
			repsContent := e.Find("div.rep > div.reptext").Text()
			edAttr = strings.ReplaceAll(strings.ReplaceAll(edAttr, "cmn.tool(this,", ""), "); return false;", "")
			repsDate := postjung.CommentDate{}
			json.Unmarshal([]byte(edAttr), &repsDate)
			log.Println(cyan("reps"), repsDate, repsContent)

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

		log.Println(cyan("comment"), author, comment, commentDate)

	})
	defer func() {
		res.Body.Close()
	}()
}

func getContentsFromTopic(ch <-chan scrapTopic, done chan bool, roomID string) {
	for c := range ch {
		c := c
		log.Println(red("Link is: "), c)
		getContent := postjung.Scraping{colly.NewCollector()}
		getContent.Scraping(c.link, "body", ".mainbox", func(_ int, elem *colly.HTMLElement) {
			title := emoji.RemoveAll(elem.ChildText("body > div.mainbox > h1"))
			author := emoji.RemoveAll(elem.ChildText("#maincontent > #hbar1 > a:nth-child(2)"))
			allContent := emoji.RemoveAll(strings.ReplaceAll(elem.ChildText("#maincontent"), "'", ""))
			imgs := elem.ChildAttrs("img", "src")
			imgsJSON, _ := json.MarshalIndent(imgs, "", "    ")
			tags := []string{}
			elem.ForEach("#maincontent > div.sptags > a", func(_ int, t *colly.HTMLElement) {
				// log.Println(cyan("t"), t.Text)
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
			//////////Log//////////
			/*
				log.Println(red("title"), title)
				log.Println(red("author"), author)
				log.Println(red("from"), allContent)
				log.Println(red("img"), string(imgsJSON))
				log.Println(cyan("tags"), tagsStore)
				log.Println(red("publishDate"), publishDate)
				log.Println(cyan("comment"), comments)
				log.Println(cyan("messageType"), messageType)
				log.Println(cyan("websiteType"), websiteType)
				log.Println(cyan("roomID"), roomID)
			*/
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
			// log.Println(cyan("content"), content)
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
	forumPage.Scraping(categoriesURL, "body > div:nth-child(9)", "a.xnav", func(_ int, elem *colly.HTMLElement) {
		if elem.Text == "next >" {
			link := elem.Attr("href")
			log.Println(red("current "), categoriesURL)
			log.Println(red("next "), link)
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
		log.Println(green("link "), postjung.SiteConfig["site"]+link)
		log.Println(green("room"), elem.Text)
		roomID := strings.Join(strings.Split(link, "-")[1:2], "")
		log.Println(green("RoomID "), roomID)
		i := 0
		var chk bool
		chk = true
		for chk {
			findAllTopicPage(postjung.SiteConfig["site"]+"board.php?id="+roomID+"&page="+strconv.Itoa(i), &chk)
			i++
		}
		log.Println(green("Total "), strconv.Itoa(i))
		RID, err := strconv.Atoi(roomID)
		if err != nil {
			log.Fatal(red("error :"), err)
		}
		forum := postjung.Forum{
			RoomName:  elem.Text,
			TotalPage: i,
			RoomId:    RID,
		}
		result := db.Create(&forum)
		log.Println("result", result)
	})
}

func findAllTopicPage(link string, chk *bool) {
	topicPage := postjung.Scraping{colly.NewCollector()}
	topicPage.Scraping(link, "div.pagebar", "a.xnav", func(_ int, elem *colly.HTMLElement) {
		if elem.Text == "next >" {
			page := elem.Attr("href")
			log.Println(red("Page of Topic "), postjung.SiteConfig["site"]+page[1:])
			*chk = true
		} else {
			*chk = false
		}
	})
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
