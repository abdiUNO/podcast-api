package podcasts

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
	"go-podcast-api/database/orm"
	"go-podcast-api/utils"
	"strconv"
	"time"
)

var redisClient = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

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

func ParseFeed(feedUrl string) (*gofeed.Feed, error) {

	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(feedUrl)

	_, err := json.Marshal(feed)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	//err = client.Set(feedUrl, out, 0).Err()
	//if err != nil {
	//	log.Fatal(err)
	//	return nil
	//}

	return feed, nil
}

func FindPodcast(podId string) (*Podcast, *gofeed.Feed, *utils.Error) {
	podcast := &Podcast{}
	err := GetDB().Table("podcasts").First(&podcast, "id = ?", podId).Error

	if err != nil {
		log.Println(err)
		if err == gorm.ErrRecordNotFound {
			return &Podcast{}, &gofeed.Feed{}, utils.NewError(utils.ENOTFOUND, "Conversation not found", nil)
		}
		return &Podcast{}, &gofeed.Feed{}, utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	var cachedFeed, e = utils.GetFromCache(redisClient, podcast.FeedUrl)

	if e != nil {
		log.Println(err)
		dbFeed := &gofeed.Feed{}

		if dbFeed, err = ParseFeed(podcast.FeedUrl); err != nil {
			return podcast, dbFeed, utils.NewError(utils.EINTERNAL, "internal database error", err)
		}

		log.Println("Retrieving Podcast Feed from DB")

		if ok := utils.SetInCache(redisClient, podcast.FeedUrl, dbFeed, 15*time.Minute); !ok {
			return podcast, dbFeed, utils.NewError(utils.EINTERNAL, "internal database error", err)
		}

		log.Println("Set Podcast Feed in Redis Cache")

		return podcast, dbFeed, nil
	}

	log.Println("Retrieving Podcast Feed from Redis Cache")

	feed := &gofeed.Feed{}
	err = json.Unmarshal(cachedFeed, &feed)

	if err != nil {
		log.Println(err)
		return podcast, feed, utils.NewError(utils.EINTERNAL, "internal database error", err)
	}

	return podcast, feed, nil
}

func FindAllPodcasts() (*[]Podcast, *utils.Error) {

	var cachedPodcasts, err = utils.GetFromCache(redisClient, "AllPodcasts")

	if err != nil {
		log.Println(err)
		dbProducts := &[]Podcast{}
		if err := GetDB().Limit(10).Table("podcasts").Find(&dbProducts).Error; err != nil {
			return dbProducts, utils.NewError(utils.EINTERNAL, "internal database error", err)
		}

		log.Println("Retrieving AllPodcasts from DB")

		if ok := utils.SetInCache(redisClient, "AllPodcasts", dbProducts, 45*time.Minute); !ok {
			return dbProducts, utils.NewError(utils.EINTERNAL, "internal database error", err)
		}

		log.Println("Set AllPodcasts in Redis Cache")

		return dbProducts, nil
	}

	log.Println("Retrieving AllPodcasts from Redis Cache")

	podcasts := &[]Podcast{}
	err = json.Unmarshal(cachedPodcasts, &podcasts)

	if err != nil {
		log.Println(err)
		return nil, utils.NewError(utils.EINTERNAL, "internal database error", err)
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

func QueryPodcast(query string) (*[]Podcast, *utils.Error) {
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
