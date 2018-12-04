package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

func isMySQLDuplicatedRecordErr(err error) bool {
	merr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}
	return merr.Number == 1062
}

func StaticMiddleware(config middleware.StaticConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Root == "" {
		config.Root = "." // For security we want to restrict to CWD.
	}
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultStaticConfig.Skipper
	}
	if config.Index == "" {
		config.Index = middleware.DefaultStaticConfig.Index
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}

			p := c.Request().URL.Path
			if strings.HasSuffix(c.Path(), "*") { // When serving from a group, e.g. `/static*`.
				p = c.Param("*")
			}
			p, err = url.PathUnescape(p)
			if err != nil {
				return
			}
			target := path.Clean("/" + p)
			switch target {
			case "/", "/index.html", "/service-worker.js":
				c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
			}
			name := filepath.Join(config.Root, target)

			fi, err := os.Stat(name)
			if err != nil {
				if os.IsNotExist(err) {
					if err = next(c); err != nil {
						if he, ok := err.(*echo.HTTPError); ok {
							if config.HTML5 && he.Code == http.StatusNotFound {
								return c.File(filepath.Join(config.Root, config.Index))
							}
						}
						return
					}
				}
				return
			}

			if fi.IsDir() {
				index := filepath.Join(name, config.Index)
				fi, err = os.Stat(index)

				if err != nil {
					if os.IsNotExist(err) {
						return next(c)
					}
					return
				}

				return c.File(index)
			}

			return c.File(name)
		}
	}
}
