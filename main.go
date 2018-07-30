package main

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

var (
	db *gorm.DB
)

func init() {
	viper.SetDefault("MYSQL_HOSTNAME", "localhost")
	viper.SetDefault("MYSQL_DATABASE", "cmkp")
	viper.SetDefault("MYSQL_USERNAME", "root")
	viper.SetDefault("MYSQL_PASSWORD", "password")
	viper.SetDefault("PORT", 3000)
	viper.SetDefault("CERT_CACHE_DIR", "./cert-cache")
	viper.SetDefault("DOMAIN", "")
	viper.AutomaticEnv()
}

func main() {
	// connect database
	var err error
	db, err = gorm.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true",
		viper.GetString("MYSQL_USERNAME"),
		viper.GetString("MYSQL_PASSWORD"),
		viper.GetString("MYSQL_HOSTNAME"),
		viper.GetString("MYSQL_DATABASE"),
	))
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").AutoMigrate(tables...).Error; err != nil {
		log.Fatal(err)
	}

	// init db data
	if err = initAdminUser(); err != nil {
		log.Fatal(err)
	}
	if err = initDeadline(); err != nil {
		log.Fatal(err)
	}

	// init web server
	e := echo.New()
	e.Validator = newValidator()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 2}))
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/api")
		},
		Root:   "static",
		Index:  "index.html",
		Browse: false,
		HTML5:  true,
	}))

	// routing api
	require := accessControlMiddleware
	e.POST("/api/login", login)
	api := e.Group("/api", middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:     &JwtCustomClaims{},
		SigningKey: []byte(jwtSecret),
	}), require(LevelUser))
	{
		me := api.Group("/me")
		me.GET("", getMe)
		me.PATCH("/password", changeMyPassword)
		me.GET("/requests", getMyRequests)
		me.POST("/requests", createMyRequest)
		me.GET("/request-notes", getMyRequestNotes)
		me.POST("/request-notes", postMyRequestNotes)
		me.GET("/circle-priorities", getMyCirclePriorities)
		me.POST("/circle-priorities", updateMyCirclePriorities)

		users := api.Group("/users")
		users.GET("", getUsers)
		users.POST("", postUsers, require(LevelAdmin))
		users.GET("/:id", getUser)
		users.PATCH("/:id/entry", changeUserEntry, require(LevelAdmin))
		users.PATCH("/:id/password", changeUserPassword, require(LevelAdmin))
		users.PATCH("/:id/permission", changeUserPermission, require(LevelAdmin))
		users.GET("/:id/requests", getUserRequests, require(LevelPlanner))
		users.GET("/:id/request-notes", getUserRequestNotes, require(LevelPlanner))
		users.GET("/:id/circle-priorities", getUserCirclePriorities, require(LevelPlanner))

		circles := api.Group("/circles")
		circles.GET("", getCircles)
		circles.GET("/requested", getRequestedCircles)
		circles.GET("/:id", getCircle)
		circles.GET("/:id/memos", getCircleMemos)
		circles.POST("/:id/memos", postCircleMemo)
		circles.GET("/:id/items", getCircleItems)

		circleMemo := api.Group("/circle-memos")
		circleMemo.GET("/:mid", getCircleMemo)
		circleMemo.DELETE("/:mid", deleteCircleMemo)

		items := api.Group("/items")
		items.POST("", postItem)
		items.GET("/:id", getItem)
		items.PATCH("/:id/name", updateItemName, require(LevelPlanner))
		items.PATCH("/:id/price", updateItemPrice)

		requests := api.Group("/requests")
		requests.GET("", getRequests, require(LevelPlanner))
		requests.POST("", postRequest, require(LevelPlanner))
		requests.GET("/:id", getRequest)
		requests.PATCH("/:id", editRequest)
		requests.DELETE("/:id", deleteRequest)

		requestNotes := api.Group("/request-notes")
		requestNotes.GET("", getRequestNotes, require(LevelPlanner))
		requestNotes.POST("", postRequestNote)
		requestNotes.GET("/:id", getRequestNote)
		requestNotes.DELETE("/:id", deleteRequestNote)

		/*
			assignmentNotes := api.Group("/assignment-notes")
			assignmentNotes.POST("", postAssignmentNote, require(LevelPlanner))
			assignmentNotes.GET("/:id", getAssignmentNote)
			assignmentNotes.DELETE("/:id", deleteAssignmentNote, require(LevelPlanner))
		*/

		deadlines := api.Group("/deadlines")
		deadlines.GET("", getDeadLines)
		deadlines.PUT("", setDeadLines, require(LevelPlanner))
	}

	// start server
	go func() {
		if len(viper.GetString("DOMAIN")) > 0 && viper.GetInt("PORT") == 443 {
			e.AutoTLSManager.Cache = autocert.DirCache(viper.GetString("CERT_CACHE_DIR"))
			e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(viper.GetString("DOMAIN"))
			if err := e.StartAutoTLS(":443"); err != nil {
				e.Logger.Info("shutting down the server")
			}
		} else {
			if err := e.Start(fmt.Sprintf(":%d", viper.GetInt("PORT"))); err != nil {
				e.Logger.Info("shutting down the server")
			}
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func initAdminUser() error {
	u := &User{}
	if err := db.Where(&User{Permission: LevelAdmin}).Take(u).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			salt := generateRandomString()
			u = &User{
				Name:              "admin",
				DisplayName:       "サーバー管理人",
				Salt:              salt,
				EncryptedPassword: toHash("password", salt),
				Permission:        LevelAdmin,
			}
			if err := db.Create(u).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func initDeadline() error {
	for i := 0; i < 4; i++ {
		if _, ok := getDayDeadLine(i); !ok {
			if err := setDayDeadLine(i, time.Now().Truncate(time.Minute)); err != nil {
				return err
			}
		}
	}
	return nil
}
