package model

import (
	"context"
	"time"
)

type Deadline struct {
	ID        int       `gorm:"primary_key"`
	Day       int       `gorm:"unique"`
	Datetime  time.Time `sql:"DEFAULT:current_timestamp"`
	UpdatedAt time.Time `gorm:"precision:6"`
}

func (dl *Deadline) IsOver() bool {
	return dl.Datetime.Before(time.Now())
}

func GetDeadline(ctx context.Context, day int) (*Deadline, error) {
	v := Deadline{}
	if err := orm(ctx).Where("day = ?", day).First(&v).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}
	return &v, nil
}

func GetDeadlines(ctx context.Context) ([]*Deadline, error) {
	arr := make([]*Deadline, 0)
	if err := orm(ctx).Find(&arr).Error; err != nil {
		panic(err)
	}
	return arr, nil
}

func SetDeadline(ctx context.Context, day int, time time.Time) (*Deadline, error) {
	db := orm(ctx)
	dl := &Deadline{}
	if err := db.Where(map[string]interface{}{"day": day}).FirstOrCreate(dl).Error; err != nil {
		panic(err)
	}

	dl.Datetime = time
	if err := db.Save(dl).Error; err != nil {
		return nil, err
	}
	return dl, nil
}
