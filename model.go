package main

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

const (
	LevelUser = iota
	LevelPlanner
	LevelAdmin
)

var (
	tables = []interface{}{
		&User{},
		&Circle{},
		&CircleMemo{},
		&Item{},
		&UserRequestItem{},
		&UserRequestNote{},
		&UserCirclePriority{},
		&DeadLine{},
		// &AssignmentGroup{},
		// &AssignmentGroupNote{},
		// &AssignmentItem{},
		// &FreeAssignmentItem{},
		// &UserRequestNoteUnread{},
	}
)

type JwtCustomClaims struct {
	jwt.StandardClaims
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

func (jcc *JwtCustomClaims) GetUID() uint {
	return mustParseUint(jcc.Subject)
}

type User struct {
	ID                uint      `gorm:"primary_key"             json:"id"`
	Name              string    `gorm:"type:varchar(20);unique" json:"name"`
	DisplayName       string    `gorm:"type:varchar(30)"        json:"display_name"`
	EncryptedPassword string    `gorm:"type:text"               json:"-"`
	Salt              string    `gorm:"type:text"               json:"-"`
	Permission        int       `                               json:"permission"`
	EntryDay1         bool      `                               json:"entry_day1"`
	EntryDay2         bool      `                               json:"entry_day2"`
	EntryDay3         bool      `                               json:"entry_day3"`
	CreatedAt         time.Time `gorm:"precision:6"             json:"created_at"`
	UpdatedAt         time.Time `gorm:"precision:6"             json:"updated_at"`
}

// クライアントでフルキャッシュ
type Circle struct {
	ID           uint          `gorm:"primary_key"      json:"id"`
	CatalogID    *int          `                        json:"-"`
	Name         string        `                        json:"name"`
	Author       string        `                        json:"author"`
	Hall         string        `gorm:"size:1"           json:"hall"`
	Day          int           `                        json:"day"`
	Block        string        `gorm:"size:1"           json:"block"`
	Space        string        `                        json:"space"`
	LocationType int           `                        json:"location_type"`
	Genre        string        `                        json:"genre"`
	PixivID      *int          `                        json:"pixiv_id"`
	TwitterID    *string       `gorm:"type:varchar(15)" json:"twitter_id"`
	NiconicoID   *int          `                        json:"niconico_id"`
	Website      string        `gorm:"type:text"        json:"website"`
	Memos        []*CircleMemo `gorm:"association_autoupdate:false;association_autocreate:false" json:"-"`
	UpdatedAt    time.Time     `gorm:"precision:6"      json:"updated_at"`
}

// クライアントでフルキャッシュ
type CircleMemo struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CircleID  uint       `gorm:"index"       json:"circle_id"`
	UserID    uint       `                   json:"user_id"`
	Memo      string     `gorm:"type:text"   json:"content"`
	CreatedAt time.Time  `gorm:"precision:6" json:"created_at"`
	DeletedAt *time.Time `gorm:"precision:6" json:"-"`
}

type Item struct {
	ID        uint      `gorm:"primary_key"                                               json:"id"`
	CircleID  uint      `gorm:"unique_index:circle_item_name"                             json:"circle_id"`
	Circle    *Circle   `gorm:"association_autoupdate:false;association_autocreate:false" json:"-"`
	Name      string    `gorm:"type:varchar(100);unique_index:circle_item_name"           json:"name"`
	Price     int       `                                                                 json:"price"`
	CreatedAt time.Time `gorm:"precision:6"                                               json:"created_at"`
	UpdatedAt time.Time `gorm:"precision:6"                                               json:"updated_at"`
}

type UserRequestItem struct {
	ID        uint      `gorm:"primary_key"                                               json:"id"`
	UserID    uint      `gorm:"unique_index:user_item"                                    json:"user_id"`
	ItemID    uint      `gorm:"unique_index:user_item"                                    json:"item_id"`
	Item      *Item     `gorm:"association_autoupdate:false;association_autocreate:false" json:"item,omitempty"`
	Num       int       `                                                                 json:"num"`
	CreatedAt time.Time `gorm:"precision:6"                                               json:"created_at"`
	UpdatedAt time.Time `gorm:"precision:6"                                               json:"updated_at"`
}

// クライアントでフルキャッシュ
type UserRequestNote struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	UserID    uint       `                   json:"user_id"`
	Content   string     `gorm:"type:text"   json:"content"`
	CreatedAt time.Time  `gorm:"precision:6" json:"created_at"`
	DeletedAt *time.Time `gorm:"precision:6" json:"-"`
}

type UserRequestNoteUnread struct {
	UserID uint `gorm:"unique_index:user_note" json:"user_id"`
	NoteID uint `gorm:"unique_index:user_note" json:"note_id"`
}

type UserCirclePriority struct {
	ID        uint      `gorm:"primary_key"           json:"id"`
	UserID    uint      `gorm:"unique_index:user_day" json:"user_id"`
	Day       int       `gorm:"unique_index:user_day" json:"day"`
	Priority1 *uint     `                             json:"priority1"`
	Priority2 *uint     `                             json:"priority2"`
	Priority3 *uint     `                             json:"priority3"`
	Priority4 *uint     `                             json:"priority4"`
	Priority5 *uint     `                             json:"priority5"`
	CreatedAt time.Time `gorm:"precision:6"           json:"created_at"`
	UpdatedAt time.Time `gorm:"precision:6"           json:"updated_at"`
}

func (ucp *UserCirclePriority) GetPrioritySlice() []uint {
	var priority []uint
	if ucp.Priority1 != nil {
		priority = append(priority, *ucp.Priority1)
		if ucp.Priority2 != nil {
			priority = append(priority, *ucp.Priority2)
			if ucp.Priority3 != nil {
				priority = append(priority, *ucp.Priority3)
				if ucp.Priority4 != nil {
					priority = append(priority, *ucp.Priority4)
					if ucp.Priority5 != nil {
						priority = append(priority, *ucp.Priority5)
					}
				}
			}
		}
	}
	return priority
}

type AssignmentGroup struct {
	ID        uint `gorm:"primary_key"`
	Day       int
	UserID    *uint
	CreatedAt time.Time  `gorm:"precision:6"`
	UpdatedAt time.Time  `gorm:"precision:6"`
	DeletedAt *time.Time `gorm:"precision:6"`
}

type AssignmentGroupNote struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	WriterID  uint       `                   json:"writer_id"`
	GroupID   uint       `                   json:"group_id"`
	Content   string     `gorm:"type:text"   json:"content"`
	CreatedAt time.Time  `gorm:"precision:6" json:"created_at"`
	DeletedAt *time.Time `gorm:"precision:6" json:"-"`
}

type AssignmentItem struct {
	ID        uint `gorm:"primary_key"`
	GroupID   uint
	ItemID    uint
	Item      *Item `gorm:"association_autoupdate:false;association_autocreate:false" json:"-"`
	Num       int
	Order     int
	CreatedAt time.Time  `gorm:"precision:6"`
	DeletedAt *time.Time `gorm:"precision:6"`
}

type FreeAssignmentItem struct {
	ID        uint `gorm:"primary_key"`
	Day       int
	ItemID    uint
	Item      *Item `gorm:"association_autoupdate:false;association_autocreate:false" json:"-"`
	Num       int
	ChargerID *uint
	CreatedAt time.Time `gorm:"precision:6"`
	UpdatedAt time.Time `gorm:"precision:6"`
}

type DeadLine struct {
	ID       uint   `gorm:"primary_key"`
	Key      string `gorm:"unique"`
	DateTime time.Time
}
