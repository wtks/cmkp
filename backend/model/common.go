package model

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
	"regexp"
)

const (
	ORMKey     = "orm"
	LoadersKey = "loaders"
)

var (
	v      *validator.Validate
	tables = []interface{}{
		&User{},
		&Circle{},
		&CircleMemo{},
		&Item{},
		&UserRequestItem{},
		&UserRequestNote{},
		&Deadline{},
		&UserCirclePriority{},
		&Content{},
	}
	usernameRegex = regexp.MustCompile("^[0-9a-zA-Z_-]{1,20}$")
	ErrForbidden  = errors.New("forbidden")
)

type Loaders struct {
	Circle          *CircleLoader
	CircleMemo      *CircleMemoLoader
	Item            *ItemLoader
	User            *UserLoader
	UserRequestItem *UserRequestItemLoader
	UserRequestNote *UserRequestNoteLoader
}

func init() {
	v = validator.New()
	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return usernameRegex.MatchString(fl.Field().String())
	})
}

func InitTables(ctx context.Context) error {
	return orm(ctx).Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").AutoMigrate(tables...).Error
}

func orm(ctx context.Context) *gorm.DB {
	return ctx.Value(ORMKey).(*gorm.DB)
}

func getCtxUserId(ctx context.Context) int {
	return ctx.Value("userId").(int)
}

func getCtxUserRole(ctx context.Context) Role {
	return ctx.Value("role").(Role)
}

func IsGranted(ctx context.Context, role Role, require Role) bool {
	switch require {
	case RoleUser:
		return true
	case RolePlanner:
		return role == RoleAdmin || role == RolePlanner
	case RoleAdmin:
		return role == RoleAdmin
	default:
		return false
	}
}

func CreateLoaders(ctx context.Context) *Loaders {
	return &Loaders{
		Circle:          createCircleLoader(ctx),
		CircleMemo:      createCircleMemoLoader(ctx),
		Item:            createItemLoader(ctx),
		User:            createUserLoader(ctx),
		UserRequestItem: createUserRequestItemLoader(ctx),
		UserRequestNote: createUserRequestNoteLoader(ctx),
	}
}

func isMySQLDuplicatedRecordErr(err error) bool {
	merr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}
	return merr.Number == 1062
}

func panicUnlessNotFound(err error) error {
	if gorm.IsRecordNotFoundError(err) {
		return err
	}
	panic(err)
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), err
}
