package model

import (
	"context"
	"github.com/jinzhu/gorm"
	"time"
)

func getUserRequestNoteLoader(ctx context.Context) (*UserRequestNoteLoader, bool) {
	v, e := ctx.Value(LoadersKey).(*Loaders)
	return v.UserRequestNote, e
}

func createUserRequestNoteLoader(ctx context.Context) *UserRequestNoteLoader {
	db := orm(ctx)

	return &UserRequestNoteLoader{
		wait:     1 * time.Millisecond,
		maxBatch: 1000,
		fetch: func(keys []int) (items []*UserRequestNote, errors []error) {
			items = make([]*UserRequestNote, len(keys))
			errors = make([]error, len(keys))

			var tmp []*UserRequestNote
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

type UserRequestNote struct {
	ID        int        `gorm:"primary_key" json:"id"`
	UserID    int        `                   json:"userId"`
	Content   string     `gorm:"type:text"   json:"content"`
	CreatedAt time.Time  `gorm:"precision:6" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"precision:6" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"precision:6" json:"-"`
}

func (urn *UserRequestNote) User(ctx context.Context) (*User, error) {
	return GetUserByID(ctx, urn.UserID)
}

func GetUserRequestNoteByID(ctx context.Context, id int) (*UserRequestNote, error) {
	if loader, ok := getUserRequestNoteLoader(ctx); ok {
		return loader.Load(id)
	}

	v := UserRequestNote{}
	if err := orm(ctx).First(&v, id).Error; err != nil {
		panic(err)
	}
	return &v, nil
}

func GetUserRequestNotes(ctx context.Context) ([]*UserRequestNote, error) {
	if loader, ok := getUserRequestNoteLoader(ctx); ok {
		ids := make([]int, 0)
		if err := orm(ctx).Model(UserRequestNote{}).Order("updated_at DESC").Pluck("id", &ids).Error; err != nil {
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

	arr := make([]*UserRequestNote, 0)
	if err := orm(ctx).Order("updated_at DESC").Find(&arr).Error; err != nil {
		panic(err)
	}
	return arr, nil

}

func GetUserRequestNotesByUserID(ctx context.Context, userID int) ([]*UserRequestNote, error) {
	cond := &UserRequestNote{UserID: userID}
	if loader, ok := getUserRequestNoteLoader(ctx); ok {
		ids := make([]int, 0)
		if err := orm(ctx).Model(UserRequestNote{}).Where(cond).Order("updated_at DESC").Pluck("id", &ids).Error; err != nil {
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

	arr := make([]*UserRequestNote, 0)
	if err := orm(ctx).Order("updated_at DESC").Find(&arr, cond).Error; err != nil {
		panic(err)
	}
	return arr, nil
}

func CreateUserRequestNote(ctx context.Context, userID int, content string) (*UserRequestNote, error) {
	if err := v.Var(content, "required"); err != nil {
		return nil, err
	}

	urn := &UserRequestNote{
		UserID:  userID,
		Content: content,
	}
	if err := orm(ctx).Create(urn).Error; err != nil {
		return nil, err
	}

	return urn, nil
}

func DeleteUserRequestNote(ctx context.Context, id int) error {
	return orm(ctx).Delete(&UserRequestNote{ID: id}).Error
}
