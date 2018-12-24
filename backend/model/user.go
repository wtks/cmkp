package model

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func getUserLoader(ctx context.Context) (*UserLoader, bool) {
	v, e := ctx.Value(LoadersKey).(*Loaders)
	return v.User, e
}

func createUserLoader(ctx context.Context) *UserLoader {
	db := orm(ctx)

	return &UserLoader{
		wait:     1 * time.Millisecond,
		maxBatch: 1000,
		fetch: func(keys []int) (items []*User, errors []error) {
			items = make([]*User, len(keys))
			errors = make([]error, len(keys))

			var tmp []*User
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

type User struct {
	ID                int    `gorm:"primary_key"`
	Name              string `gorm:"type:varchar(20);unique"`
	DisplayName       string `gorm:"type:varchar(30)"`
	EncryptedPassword string `gorm:"type:text"`
	Role              Role   `gorm:"type:varchar(10)"`
	EntryDay1         bool
	EntryDay2         bool
	EntryDay3         bool
	CreatedAt         time.Time `gorm:"precision:6"`
	UpdatedAt         time.Time `gorm:"precision:6"`
}

func (u *User) Entry(day int) bool {
	switch day {
	case 1:
		return u.EntryDay1
	case 2:
		return u.EntryDay2
	case 3:
		return u.EntryDay3
	default:
		return false
	}
}

func (u *User) EntryDays() []int {
	result := make([]int, 0)
	for i := 1; i <= 3; i++ {
		if u.Entry(i) {
			result = append(result, i)
		}
	}
	return result
}

func (u *User) RequestItems(ctx context.Context) ([]*UserRequestItem, error) {
	if getCtxUserId(ctx) == u.ID || IsGranted(ctx, getCtxUserRole(ctx), RolePlanner) {
		return GetUserRequestItemsByUserID(ctx, u.ID)
	} else {
		return nil, ErrForbidden
	}
}

func (u *User) RequestNotes(ctx context.Context) ([]*UserRequestNote, error) {
	if getCtxUserId(ctx) == u.ID || IsGranted(ctx, getCtxUserRole(ctx), RolePlanner) {
		return GetUserRequestNotesByUserID(ctx, u.ID)
	} else {
		return nil, ErrForbidden
	}
}

func (u *User) CirclePriorities(ctx context.Context) ([]*UserCirclePriority, error) {
	if getCtxUserId(ctx) == u.ID || IsGranted(ctx, getCtxUserRole(ctx), RolePlanner) {
		return GetUserCirclePriorityByUserID(ctx, u.ID)
	} else {
		return nil, ErrForbidden
	}
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(password))
	return err == nil
}

func GetUsers(ctx context.Context) ([]*User, error) {
	if loader, ok := getUserLoader(ctx); ok {
		ids := make([]int, 0)
		if err := orm(ctx).Model(User{}).Pluck("id", &ids).Error; err != nil {
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

	arr := make([]*User, 0)
	if err := orm(ctx).Find(&arr).Error; err != nil {
		panic(err)
	}
	return arr, nil
}

func GetUserByID(ctx context.Context, id int) (*User, error) {
	if loader, ok := getUserLoader(ctx); ok {
		return loader.Load(id)
	}

	user := User{}
	if err := orm(ctx).First(&user, id).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}
	return &user, nil
}

func GetRequestingUserByCircleID(ctx context.Context, id int) ([]*User, error) {
	query := orm(ctx).
		Model(User{}).
		Joins("INNER JOIN (SELECT user_request_items.user_id as id FROM items INNER JOIN user_request_items ON items.id = user_request_items.item_id WHERE items.circle_id = ? GROUP BY id) t ON users.id = t.id", id)

	if loader, ok := getUserLoader(ctx); ok {
		ids := make([]int, 0)
		if err := query.Pluck("users.id", &ids).Error; err != nil {
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

	arr := make([]*User, 0)
	if err := query.Find(&arr).Error; err != nil {
		panic(err)
	}
	return arr, nil
}

func CreateUser(ctx context.Context, name, displayName, password string) (*User, error) {
	if err := v.Var(name, "username,required"); err != nil {
		return nil, err
	}
	if err := v.Var(displayName, "max=30,required"); err != nil {
		return nil, err
	}
	if err := v.Var(password, "printascii,required,max=50"); err != nil {
		return nil, err
	}

	// check username
	u := &User{}
	if err := orm(ctx).First(u, &User{Name: name}).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			panic(err)
		}
		// OK
	} else {
		// NG
		return nil, errors.New("conflict")
	}

	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	u = &User{
		Name:              name,
		DisplayName:       displayName,
		EncryptedPassword: hash,
		Role:              RoleUser,
	}
	if err := orm(ctx).Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func ChangeUserPassword(ctx context.Context, userID int, password string) error {
	if err := v.Var(password, "printascii,required,max=50"); err != nil {
		return err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	if err := orm(ctx).Model(User{}).Where(&User{ID: userID}).Updates(&User{EncryptedPassword: hash}).Error; err != nil {
		return err
	}
	return nil
}

func ChangeUserRole(ctx context.Context, userID int, role Role) (*User, error) {
	user := User{}
	if err := orm(ctx).First(&user, userID).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}

	user.Role = role
	if err := orm(ctx).Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func ChangeUserEntry(ctx context.Context, userID, day int, entry bool) (*User, error) {
	user := User{}
	if err := orm(ctx).First(&user, userID).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}

	switch day {
	case 1:
		user.EntryDay1 = entry
	case 2:
		user.EntryDay2 = entry
	case 3:
		user.EntryDay3 = entry
	}

	if err := orm(ctx).Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func ChangeUserEntries(ctx context.Context, userID int, entries []int) (*User, error) {
	user := User{}
	if err := orm(ctx).First(&user, userID).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}

	user.EntryDay1 = false
	user.EntryDay2 = false
	user.EntryDay3 = false
	for _, v := range entries {
		switch v {
		case 1:
			user.EntryDay1 = true
		case 2:
			user.EntryDay2 = true
		case 3:
			user.EntryDay3 = true
		}
	}

	if err := orm(ctx).Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
