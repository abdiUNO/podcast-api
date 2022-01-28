package podcasts

import (
	"github.com/jinzhu/gorm"
	"go-podcast-api/database/orm"
)

type Genre struct {
	orm.GormModel
	Name string `gorm:"unique" json:"name"`
	Url  string `json:"url"`
}

func (genre *Genre) Create() error {
	err := GetDB().Create(&genre).Error
	return err
}

func GetGenreByName(name string) *Genre {
	genre := &Genre{}
	err := GetDB().Table("genres").Where("name = ?", name).First(genre).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil
	}

	return genre
}
