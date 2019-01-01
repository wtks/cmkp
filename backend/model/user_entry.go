package model

import (
	"context"
	"github.com/jinzhu/gorm"
	"time"
)

type UserEntry struct {
	ID        int       `gorm:"primary_key"`
	UserID    int       `gorm:"unique_index:user_day"`
	Day       int       `gorm:"unique_index:user_day"`
	CreatedAt time.Time `gorm:"precision:6"`
}

func GetUserEntries(ctx context.Context, userID int) ([]*UserEntry, error) {
	entries := make([]*UserEntry, 0)
	if err := orm(ctx).Where(&UserEntry{UserID: userID}).Order("day").Find(&entries).Error; err != nil {
		panic(err)
	}
	return entries, nil
}

func ChangeUserEntry(ctx context.Context, userID, day int, entry bool) (*User, error) {
	user := User{}
	if err := orm(ctx).First(&user, userID).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}

	if entry {
		if err := orm(ctx).Where(map[string]interface{}{"user_id": userID, "day": day}).FirstOrCreate(&UserEntry{}).Error; err != nil {
			return nil, err
		}
	} else {
		if err := orm(ctx).Where(map[string]interface{}{"user_id": userID, "day": day}).Delete(&UserEntry{}).Error; err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func ChangeUserEntries(ctx context.Context, userID int, entries []int) (*User, error) {
	user := User{}
	if err := orm(ctx).First(&user, userID).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}

	return &user, transact(orm(ctx), func(tx *gorm.DB) error {
		if err := tx.Where(&UserEntry{UserID: userID}).Delete(&UserEntry{}).Error; err != nil {
			return err
		}
		for _, v := range entries {
			if err := tx.Create(&UserEntry{
				UserID: userID,
				Day:    v,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
