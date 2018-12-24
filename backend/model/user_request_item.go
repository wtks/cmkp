package model

import (
	"context"
	"github.com/jinzhu/gorm"
	"time"
)

func getUserRequestItemLoader(ctx context.Context) (*UserRequestItemLoader, bool) {
	v, e := ctx.Value(LoadersKey).(*Loaders)
	return v.UserRequestItem, e
}

func createUserRequestItemLoader(ctx context.Context) *UserRequestItemLoader {
	db := orm(ctx)

	return &UserRequestItemLoader{
		wait:     1 * time.Millisecond,
		maxBatch: 1000,
		fetch: func(keys []int) (items []*UserRequestItem, errors []error) {
			items = make([]*UserRequestItem, len(keys))
			errors = make([]error, len(keys))

			var tmp []*UserRequestItem
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

type UserRequestItem struct {
	ID        int       `gorm:"primary_key"            json:"id"`
	UserID    int       `gorm:"unique_index:user_item" json:"userId"`
	ItemID    int       `gorm:"unique_index:user_item" json:"itemId"`
	Num       int       `                              json:"num"`
	CreatedAt time.Time `gorm:"precision:6"            json:"createdAt"`
	UpdatedAt time.Time `gorm:"precision:6"            json:"updatedAt"`
}

func (uri *UserRequestItem) User(ctx context.Context) (*User, error) {
	return GetUserByID(ctx, uri.UserID)
}

func (uri *UserRequestItem) Item(ctx context.Context) (*Item, error) {
	return GetItemByID(ctx, uri.ItemID)
}

func (uri *UserRequestItem) Circle(ctx context.Context) (*Circle, error) {
	return GetCircleByItemID(ctx, uri.ItemID)
}

func GetUserRequestItemByID(ctx context.Context, id int) (*UserRequestItem, error) {
	if loader, ok := getUserRequestItemLoader(ctx); ok {
		return loader.Load(id)
	}

	v := UserRequestItem{}
	if err := orm(ctx).First(&v, id).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}
	return &v, nil
}

func GetUserRequestItemByItemAndUserID(ctx context.Context, itemID, userID int) (*UserRequestItem, error) {
	query := orm(ctx).
		Model(UserRequestItem{}).
		Where(&UserRequestItem{ItemID: itemID, UserID: userID})

	if loader, ok := getUserRequestItemLoader(ctx); ok {
		ids := make([]int, 0)
		if err := query.Pluck("id", &ids).Error; err != nil {
			panic(err)
		}

		if len(ids) != 1 {
			return nil, gorm.ErrRecordNotFound
		}

		return loader.Load(ids[0])
	}

	v := UserRequestItem{}
	if err := query.First(&v).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}
	return &v, nil
}

func GetUserRequestItemsByItemID(ctx context.Context, itemID int) ([]*UserRequestItem, error) {
	cond := &UserRequestItem{ItemID: itemID}
	if loader, ok := getUserRequestItemLoader(ctx); ok {
		ids := make([]int, 0)
		if err := orm(ctx).Model(UserRequestItem{}).Where(cond).Pluck("id", &ids).Error; err != nil {
			panic(err)
		}

		arr, errs := loader.LoadAll(ids)
		for _, v := range errs {
			if v != nil {
				return nil, v
			}
		}
		return arr, nil
	}

	arr := make([]*UserRequestItem, 0)
	if err := orm(ctx).Find(&arr, cond).Error; err != nil {
		panic(err)
	}
	return arr, nil
}

func GetUserRequestItemsByUserID(ctx context.Context, userID int) ([]*UserRequestItem, error) {
	cond := &UserRequestItem{UserID: userID}
	if loader, ok := getUserRequestItemLoader(ctx); ok {
		ids := make([]int, 0)
		if err := orm(ctx).Model(UserRequestItem{}).Where(cond).Pluck("id", &ids).Error; err != nil {
			panic(err)
		}

		arr, errs := loader.LoadAll(ids)
		for _, v := range errs {
			if v != nil {
				return nil, v
			}
		}
		return arr, nil
	}

	arr := make([]*UserRequestItem, 0)
	if err := orm(ctx).Find(&arr, cond).Error; err != nil {
		panic(err)
	}
	return arr, nil
}

func CreateUserRequestItem(ctx context.Context, userID, itemID, num int) (*UserRequestItem, error) {
	if err := v.Var(num, "min=1"); err != nil {
		return nil, err
	}

	ur := UserRequestItem{}
	if err := orm(ctx).Where(&UserRequestItem{ItemID: itemID, UserID: userID}).First(&ur).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			panic(err)
		}
	}

	ur.UserID = userID
	ur.ItemID = itemID
	ur.Num = num

	if err := orm(ctx).Save(&ur).Error; err != nil {
		return nil, err
	}

	return &ur, nil
}

func ChangeUserRequestItemNum(ctx context.Context, id int, num int) (*UserRequestItem, error) {
	if err := v.Var(num, "min=1"); err != nil {
		return nil, err
	}

	v := UserRequestItem{}
	if err := orm(ctx).First(&v, id).Error; err != nil {
		return nil, err
	}

	v.Num = num
	if err := orm(ctx).Save(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func DeleteUserRequestItem(ctx context.Context, id int) error {
	return orm(ctx).Delete(&UserRequestItem{ID: id}).Error
}
