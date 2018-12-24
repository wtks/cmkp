package model

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
	"time"
)

func getItemLoader(ctx context.Context) (*ItemLoader, bool) {
	v, e := ctx.Value(LoadersKey).(*Loaders)
	return v.Item, e
}

func createItemLoader(ctx context.Context) *ItemLoader {
	db := orm(ctx)

	return &ItemLoader{
		wait:     1 * time.Millisecond,
		maxBatch: 1000,
		fetch: func(keys []int) (items []*Item, errors []error) {
			items = make([]*Item, len(keys))
			errors = make([]error, len(keys))

			var tmp []*Item
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

type Item struct {
	ID        int       `gorm:"primary_key"                                     json:"id"`
	CircleID  int       `gorm:"unique_index:circle_item_name"                   json:"circleId"`
	Name      string    `gorm:"type:varchar(100);unique_index:circle_item_name" json:"name"`
	Price     int       `                                                       json:"price"`
	CreatedAt time.Time `gorm:"precision:6"                                     json:"createdAt"`
	UpdatedAt time.Time `gorm:"precision:6"                                     json:"updatedAt"`
}

func (i *Item) Circle(ctx context.Context) (*Circle, error) {
	return GetCircleByID(ctx, i.CircleID)
}

func (i *Item) Requests(ctx context.Context) ([]*UserRequestItem, error) {
	if IsGranted(ctx, getCtxUserRole(ctx), RolePlanner) {
		return GetUserRequestItemsByItemID(ctx, i.ID)
	}
	return nil, ErrForbidden
}

func (i *Item) MyRequest(ctx context.Context) (*UserRequestItem, error) {
	r, _ := GetUserRequestItemByItemAndUserID(ctx, i.ID, getCtxUserId(ctx))
	return r, nil
}

func (i *Item) UserRequest(ctx context.Context, userId int) (*UserRequestItem, error) {
	if userId == getCtxUserId(ctx) || IsGranted(ctx, getCtxUserRole(ctx), RolePlanner) {
		r, _ := GetUserRequestItemByItemAndUserID(ctx, i.ID, userId)
		return r, nil
	}
	return nil, ErrForbidden
}

func GetItemByID(ctx context.Context, id int) (*Item, error) {
	if loader, ok := getItemLoader(ctx); ok {
		return loader.Load(id)
	}

	item := Item{}
	if err := orm(ctx).First(&item, id).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}
	return &item, nil
}

func GetItemsByCircleID(ctx context.Context, circleID int) ([]*Item, error) {
	if circleID <= 0 {
		return make([]*Item, 0), nil
	}

	cond := &Item{CircleID: circleID}
	if loader, ok := getItemLoader(ctx); ok {
		ids := make([]int, 0)
		if err := orm(ctx).Model(Item{}).Where(cond).Pluck("id", &ids).Error; err != nil {
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

	arr := make([]*Item, 0)
	if err := orm(ctx).Find(&arr, cond).Error; err != nil {
		panic(err)
	}
	return arr, nil
}

func GetRequestedItemsByCircleID(ctx context.Context, circleID int, userID *int) ([]*Item, error) {
	if circleID <= 0 {
		return make([]*Item, 0), nil
	}

	query := orm(ctx).
		Model(Item{}).
		Where(&Item{CircleID: circleID})
	if userID == nil {
		query = query.Joins("INNER JOIN (SELECT item_id as id FROM user_request_items GROUP BY item_id) t ON t.id = items.id")
	} else {
		query = query.Joins("INNER JOIN (SELECT item_id as id FROM user_request_items WHERE user_request_items.user_id = ? GROUP BY item_id) t ON t.id = items.id", *userID)
	}

	if loader, ok := getItemLoader(ctx); ok {
		ids := make([]int, 0)
		if err := query.Pluck("items.id", &ids).Error; err != nil {
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

	arr := make([]*Item, 0)
	if err := query.Find(&arr).Error; err != nil {
		panic(err)
	}
	return arr, nil
}

func CreateItem(ctx context.Context, circleID int, name string, price int) (*Item, error) {
	if err := v.Var(name, "required,max=100"); err != nil {
		return nil, err
	}
	if err := v.Var(price, "min=-1"); err != nil {
		return nil, err
	}

	_, err := GetCircleByID(ctx, circleID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.New("存在しないサークルです")
		}
		return nil, err
	}

	item := &Item{
		CircleID: circleID,
		Name:     name,
		Price:    price,
	}
	if err := orm(ctx).Create(item).Error; err != nil {
		if isMySQLDuplicatedRecordErr(err) {
			return nil, errors.New("既に登録されている商品名です")
		}
		return nil, err
	}

	return item, nil
}

func ChangeItemName(ctx context.Context, id int, name string) (*Item, error) {
	if err := v.Var(name, "required,max=100"); err != nil {
		return nil, err
	}

	item := Item{}
	if err := orm(ctx).First(&item, id).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}

	item.Name = name
	if err := orm(ctx).Save(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func ChangeItemPrice(ctx context.Context, id int, price int) (*Item, error) {
	if err := v.Var(price, "min=-1"); err != nil {
		return nil, err
	}

	item := Item{}
	if err := orm(ctx).First(&item, id).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}

	item.Price = price
	if err := orm(ctx).Save(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
