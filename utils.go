package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/scrypt"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	jwtSecret     = ""
	usernameRegex = regexp.MustCompile("^[0-9a-zA-Z_-]{1,20}$")
)

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

func newValidator() *Validator {
	v := validator.New()
	v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return usernameRegex.MatchString(fl.Field().String())
	})

	return &Validator{validator: v}
}

func convertStringsToUints(arr []string) (result []uint) {
	for _, v := range arr {
		n, err := strconv.ParseUint(v, 10, 32)
		if err == nil {
			result = append(result, uint(n))
		}
	}
	return result
}

func mustParseUint(str string) uint {
	n, _ := strconv.ParseUint(str, 10, 32)
	return uint(n)
}

func toHash(pass string, salt string) string {
	converted, _ := scrypt.Key([]byte(pass), []byte(salt), 16384, 8, 1, 32)
	return hex.EncodeToString(converted[:])
}

func accessControlMiddleware(level int) func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Get("permission") == nil {
				claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)
				user := &User{}
				if err := db.Select("permission").First(user, mustParseUint(claim.Subject)).Error; err != nil {
					c.Logger().Error(err)
					return echo.NewHTTPError(http.StatusInternalServerError)
				}
				c.Set("permission", user.Permission)
			}

			if c.Get("permission").(int) >= level {
				return next(c)
			} else {
				return echo.NewHTTPError(http.StatusForbidden)
			}
		}
	}
}

func getPermission(c echo.Context) int {
	p, ok := c.Get("permission").(int)
	if !ok {
		return LevelUser
	}
	return p
}

func setDayDeadLine(day int, datetime time.Time) error {
	return setDeadLine(fmt.Sprintf("day%d", day), datetime)
}

func getDayDeadLine(day int) (time.Time, bool) {
	return getDeadLine(fmt.Sprintf("day%d", day))
}

func getDeadLine(key string) (time.Time, bool) {
	dl := &DeadLine{}
	if err := db.Where(&DeadLine{Key: key}).Take(dl).Error; err != nil {
		return time.Time{}, false
	}

	return dl.DateTime, true
}

func setDeadLine(key string, datetime time.Time) error {
	dl := &DeadLine{}
	if err := db.Select("id").Where(&DeadLine{Key: key}).Take(dl).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			dl.Key = key
			dl.DateTime = datetime
			return db.Create(dl).Error
		}
		return err
	}

	return db.Model(dl).Update("date_time", datetime).Error
}

func bindAndValidate(c echo.Context, v interface{}) error {
	if err := c.Bind(v); err != nil {
		return err
	}
	if err := c.Validate(v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	return nil
}

func generateRandomString() string {
	return base64.RawStdEncoding.EncodeToString(uuid.NewV4().Bytes())
}

func hasFlag(c echo.Context, flag string) bool {
	_, ok := c.QueryParams()[flag]
	return ok
}

func isMySQLDuplicatedRecordErr(err error) bool {
	merr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}
	return merr.Number == 1062
}
