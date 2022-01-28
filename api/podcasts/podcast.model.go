package podcasts

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/mmcdole/gofeed"
	"go-podcast-api/database/orm"
	"go-podcast-api/utils"
	"log"
	"strconv"
)

func GetDB() *gorm.DB {
	return orm.DBCon
}

type Podcast struct {
	orm.GormModel
	Name          string `json:"name"`
	Description   string `sql:"type:longtext"`
	ItunesId      int    `json:"itunesId"`
	PublisherName string `json:"publisherName"`
	PublisherID   int    `json:"publisherId"`
	ImageURlSM    string `json:"imageUrlSM"`
	ImageUrlMD    string `json:"imageUrlMD"`
	ImageUrlLG    string `json:"imageUrlLG"`
	ImageUrlXL    string `json:"imageUrlXL"`
	Genre         Genre  `json:"genre"`
	GenreID       string `json:"-"`
	FeedUrl       string `json:"feedUrl"`
	EpisodesCount int    `json:"episodesCount"`
}

type Rank struct {
	orm.GormModel
	Score     int
	Podcast   Podcast `gorm:"foreignkey:PodcastID"`
	PodcastID string
	ItunesId  int `gorm:"unique" json:"itunesId"`
}

func TopPodcasts() (*[]Rank, *utils.Error) {
	podcasts := &[]Rank{}
	if err := GetDB().Table("ranks").Order("score asc").Preload("Podcast").Find(&podcasts).Error; err != nil {
		log.Println(err)
		return podcasts, utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	return podcasts, nil
}

func GetRankByItunesId(itunesId string) *Rank {
	rank := &Rank{}
	trackId, _ := strconv.Atoi(itunesId)

	err := GetDB().Table("ranks").Where("itunes_id = ?", trackId).First(rank).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil
	}

	return rank
}

func (podcast *Podcast) Create() error {
	err := GetDB().Create(&podcast).Error
	return err
}

func ParseFeed(feedUrl string) *gofeed.Feed {

	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(feedUrl)

	_, err := json.Marshal(feed)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	//err = client.Set(feedUrl, out, 0).Err()
	//if err != nil {
	//	log.Fatal(err)
	//	return nil
	//}

	return feed
}

func FindPodcastById(podId string) (*Podcast, *gofeed.Feed, *utils.Error) {
	podcast := &Podcast{}
	err := GetDB().Table("podcasts").First(&podcast, "id = ?", podId).Error

	if err != nil {
		log.Println(err)
		if err == gorm.ErrRecordNotFound {
			return &Podcast{}, &gofeed.Feed{}, utils.NewError(utils.ENOTFOUND, "Conversation not found", nil)
		}
		return &Podcast{}, &gofeed.Feed{}, utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	feed := ParseFeed(podcast.FeedUrl)

	return podcast, feed, nil
}

func AllPodcasts() (*[]Podcast, *utils.Error) {
	podcasts := &[]Podcast{}
	if err := GetDB().Limit(10).Table("podcasts").Find(&podcasts).Error; err != nil {
		return podcasts, utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	return podcasts, nil
}

func GetPodcastByTrack(itunesId string) *Podcast {
	podcast := &Podcast{}
	trackId, _ := strconv.Atoi(itunesId)

	err := GetDB().Table("podcasts").Preload("Genre").Where("itunes_id = ?", trackId).First(podcast).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil
	}

	return podcast
}

func SearchPodcastByName(query string) (*[]Podcast, *utils.Error) {
	podcasts := &[]Podcast{}

	//SELECT * from podcasts where MATCH(name) AGAINST('Radio' IN NATURAL LANGUAGE MODE)
	log.Println(query)
	if err := GetDB().Raw(`
		SELECT
			*,
			MATCH(name) AGAINST (? IN BOOLEAN MODE) AS score
		FROM
			podcasts
		WHERE
			MATCH(name) AGAINST (? IN BOOLEAN MODE) > 0
	`, query, query).Scan(&podcasts).Error; err != nil {
		log.Println(err)
		return podcasts, utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	return podcasts, nil
}
