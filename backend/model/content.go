package model

import (
	"context"
	"time"
)

type Content struct {
	ID        string    `gorm:"primary_key"`
	Text      string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"precision:6"`
	UpdatedAt time.Time `gorm:"precision:6"`
}

func GetContent(ctx context.Context, id string) (*Content, error) {
	v := Content{}
	if err := orm(ctx).Where(map[string]interface{}{"id": id}).First(&v).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}
	return &v, nil
}

func SetContent(ctx context.Context, id string, text string) (*Content, error) {
	v := Content{}
	if err := orm(ctx).Where(map[string]interface{}{"id": id}).FirstOrCreate(&v).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}
	v.Text = text
	if err := orm(ctx).Save(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}
