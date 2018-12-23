package model

import (
	"context"
	"github.com/jinzhu/gorm"
	"time"
)

func getCircleMemoLoader(ctx context.Context) (*CircleMemoLoader, bool) {
	v, e := ctx.Value(LoadersKey).(*Loaders)
	return v.CircleMemo, e
}

func createCircleMemoLoader(ctx context.Context) *CircleMemoLoader {
	db := orm(ctx)

	return &CircleMemoLoader{
		wait:     1 * time.Millisecond,
		maxBatch: 1000,
		fetch: func(keys []int) (items []*CircleMemo, errors []error) {
			items = make([]*CircleMemo, len(keys))
			errors = make([]error, len(keys))

			var tmp []*CircleMemo
			if err := db.Where("id IN (?)", keys).Find(&tmp).Error; err != nil {
				panic(err)
			}

			// now, tmp is sorted by id, not keys

			for i, key := range keys {
				ok := false
				for _, v := range tmp {
					if v.ID == key {
						ok = true
						items[i] = v
						break
					}
				}
				if !ok {
					errors[i] = gorm.ErrRecordNotFound
				}
			}
			return
		},
	}
}

type CircleMemo struct {
	ID        int        `gorm:"primary_key" json:"id"`
	CircleID  int        `gorm:"index"       json:"circleId"`
	UserID    int        `                   json:"userId"`
	Content   string     `gorm:"type:text"   json:"content"`
	CreatedAt time.Time  `gorm:"precision:6" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"precision:6" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"precision:6" json:"-"`
}

func (cm *CircleMemo) Circle(ctx context.Context) (*Circle, error) {
	return GetCircleByID(ctx, cm.CircleID)
}

func (cm *CircleMemo) User(ctx context.Context) (*User, error) {
	return GetUserByID(ctx, cm.UserID)
}

func GetCircleMemoByID(ctx context.Context, id int) (*CircleMemo, error) {
	if loader, ok := getCircleMemoLoader(ctx); ok {
		return loader.Load(id)
	}

	memo := CircleMemo{}
	if err := orm(ctx).First(&memo, id).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}
	return &memo, nil
}

func GetCircleMemosByCircleID(ctx context.Context, circleID int) ([]*CircleMemo, error) {
	if loader, ok := getCircleMemoLoader(ctx); ok {
		ids := make([]int, 0)
		if err := orm(ctx).Model(CircleMemo{}).Where(&CircleMemo{CircleID: circleID}).Pluck("id", &ids).Error; err != nil {
			panic(err)
		}

		items, errs := loader.LoadAll(ids)
		for _, v := range errs {
			if v != nil {
				return nil, v
			}
		}
		return items, nil
	}

	memos := make([]*CircleMemo, 0)
	if err := orm(ctx).Find(&memos, &CircleMemo{CircleID: circleID}).Error; err != nil {
		panic(err)
	}
	return memos, nil
}

func CreateCircleMemo(ctx context.Context, circleID, userID int, content string) (*CircleMemo, error) {
	if err := v.Var(content, "required"); err != nil {
		return nil, err
	}

	if _, err := GetCircleByID(ctx, circleID); err != nil {
		return nil, err
	}

	urn := &CircleMemo{
		UserID:   userID,
		CircleID: circleID,
		Content:  content,
	}
	if err := orm(ctx).Create(urn).Error; err != nil {
		return nil, err
	}

	return urn, nil
}

func DeleteCircleMemo(ctx context.Context, id int) error {
	return orm(ctx).Delete(&CircleMemo{ID: id}).Error
}
