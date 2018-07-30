package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// POST /login
func login(c echo.Context) error {
	req := struct {
		User string `json:"username" validate:"username,required"`
		Pass string `json:"password" validate:"printascii,required"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	user := &User{}
	if err := db.Where(&User{Name: req.User}).Take(user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "ユーザー名またはパスワードが間違っています")
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if toHash(req.Pass, user.Salt) != user.EncryptedPassword {
		return echo.NewHTTPError(http.StatusBadRequest, "ユーザー名またはパスワードが間違っています")
	}

	claims := &JwtCustomClaims{
		Name:        user.Name,
		DisplayName: user.DisplayName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 20).Unix(),
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
		},
	}

	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtSecret))
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, map[string]string{"token": t})
}

// GET /me
func getMe(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)
	user := &User{}
	if err := db.First(user, claim.GetUID()).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, user)
}

// PATCH /me/password
func changeMyPassword(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)
	user := User{}
	if err := db.First(&user, claim.GetUID()).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	req := struct {
		OldPassword string `json:"old_password" validate:"printascii,required"`
		NewPassword string `json:"new_password" validate:"printascii,required"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	if toHash(req.OldPassword, user.Salt) != user.EncryptedPassword {
		return echo.NewHTTPError(http.StatusUnauthorized, "password is wrong")
	}

	salt := generateRandomString()
	encrypted := toHash(req.NewPassword, salt)

	if err := db.Model(&user).Updates(&User{Salt: salt, EncryptedPassword: encrypted}).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /me/requests
func getMyRequests(c echo.Context) error {
	type resultItem struct {
		ID        uint      `json:"id"`
		ItemID    uint      `json:"item_id"`
		Item      *Item     `json:"item"`
		Num       int       `json:"num"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	type resultCircle struct {
		CircleID uint          `json:"circle_id"`
		Items    []*resultItem `json:"items"`
	}

	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)

	var requests []*UserRequestItem
	if err := db.Preload("Item").Where(&UserRequestItem{UserID: claim.GetUID()}).Find(&requests).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	temp := map[uint]*resultCircle{}
	for _, v := range requests {
		c, ok := temp[v.Item.CircleID]
		if !ok {
			c = &resultCircle{CircleID: v.Item.CircleID}
			temp[v.Item.CircleID] = c
		}
		i := &resultItem{
			ID:        v.ID,
			ItemID:    v.ItemID,
			Item:      v.Item,
			Num:       v.Num,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
		c.Items = append(c.Items, i)
	}

	result := make([]*resultCircle, 0)
	for _, v := range temp {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CircleID < result[j].CircleID
	})

	return c.JSON(http.StatusOK, result)
}

// POST /me/requests
func createMyRequest(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)

	req := struct {
		ItemID uint `json:"item_id" validate:"required"`
		Num    int  `json:"num"   validate:"min=1"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	circle := Circle{}
	if err := db.Where("id = (?)", db.Model(Item{ID: req.ItemID}).Select("circle_id").QueryExpr()).Take(&circle).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "the item is not found")
		}
		c.Logger().Error()
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if dl, ok := getDayDeadLine(circle.Day); ok && dl.Before(time.Now()) {
		return echo.NewHTTPError(http.StatusForbidden, "既にリクエスト締め切りが過ぎています")
	}

	ur := &UserRequestItem{
		UserID: claim.GetUID(),
		ItemID: req.ItemID,
		Num:    req.Num,
	}

	err := db.
		Set("gorm:insert_option", fmt.Sprintf("ON DUPLICATE KEY UPDATE num = %d, updated_at = current_timestamp(6)", ur.Num)).
		Create(&ur).
		Error
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, map[string]uint{"id": ur.ID})
}

// GET /me/request-notes
func getMyRequestNotes(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)

	ids := make([]uint, 0)
	if err := db.Model(&UserRequestNote{}).Where(&UserRequestNote{UserID: claim.GetUID()}).Order("created_at DESC").Pluck("id", &ids).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ids)
}

// POST /me/request-notes
func postMyRequestNotes(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)

	req := struct {
		Content string `json:"content" validate:"required"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	urn := &UserRequestNote{
		UserID:  claim.GetUID(),
		Content: req.Content,
	}
	if err := db.Create(urn).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{"id": urn.ID})
}

// GET /me/circle-priorities
func getMyCirclePriorities(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)

	var ucp []*UserCirclePriority
	if err := db.Where(&UserCirclePriority{UserID: claim.GetUID()}).Find(&ucp).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	res := &struct {
		Enterprise []uint `json:"enterprise"`
		Day1       []uint `json:"day1"`
		Day2       []uint `json:"day2"`
		Day3       []uint `json:"day3"`
	}{}
	res.Enterprise = make([]uint, 0)
	res.Day1 = make([]uint, 0)
	res.Day2 = make([]uint, 0)
	res.Day3 = make([]uint, 0)
	for _, v := range ucp {
		switch v.Day {
		case 0:
			res.Enterprise = v.GetPrioritySlice()
		case 1:
			res.Day1 = v.GetPrioritySlice()
		case 2:
			res.Day2 = v.GetPrioritySlice()
		case 3:
			res.Day3 = v.GetPrioritySlice()
		default:
			continue
		}
	}

	return c.JSON(http.StatusOK, res)
}

// POST /me/circle-priorities
func updateMyCirclePriorities(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)

	req := struct {
		Day      int    `json:"day"      validate:"min=0,max=3"`
		Priority []uint `json:"priority" validate:"max=5"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	if dl, ok := getDayDeadLine(req.Day); ok && dl.Before(time.Now()) {
		return echo.NewHTTPError(http.StatusForbidden, "the day's deadline is over")
	}

	for _, v := range req.Priority {
		circle := &Circle{}
		if err := db.First(&circle, v).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid circle id: %d", v))
			}
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		if circle.Day != req.Day {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("other day's circle: %d", v))
		}
	}

	ucp := &UserCirclePriority{
		UserID: claim.GetUID(),
		Day:    req.Day,
	}

	if err := db.Where(ucp).Delete(UserCirclePriority{}).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if len(req.Priority) >= 1 {
		ucp.Priority1 = &req.Priority[0]
	}
	if len(req.Priority) >= 2 {
		ucp.Priority2 = &req.Priority[1]
	}
	if len(req.Priority) >= 3 {
		ucp.Priority3 = &req.Priority[2]
	}
	if len(req.Priority) >= 4 {
		ucp.Priority4 = &req.Priority[3]
	}
	if len(req.Priority) >= 5 {
		ucp.Priority5 = &req.Priority[4]
	}

	if err := db.Create(ucp).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /users
func getUsers(c echo.Context) error {
	var users []*User
	if err := db.Find(&users).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, users)
}

// POST /users
func postUsers(c echo.Context) error {
	req := struct {
		UserName    string `json:"username"     validate:"username,required"`
		DisplayName string `json:"display_name" validate:"max=30,required"`
		Password    string `json:"password"     validate:"printascii,required"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	// check username
	u := &User{}
	if err := db.First(u, &User{Name: req.UserName}).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		// OK
	} else {
		// NG
		return echo.NewHTTPError(http.StatusConflict, "そのIDのユーザーは既に存在します")
	}

	salt := generateRandomString()
	u = &User{
		Name:              req.UserName,
		DisplayName:       req.DisplayName,
		Salt:              salt,
		EncryptedPassword: toHash(req.Password, salt),
		Permission:        LevelUser,
	}
	if err := db.Create(u).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, u)
}

// GET /users/:id
func getUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	user := &User{}
	if err := db.First(user, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, user)
}

// PATCH /users/:id/entry
func changeUserEntry(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	user := &User{}
	if err := db.First(user, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	req := struct {
		Day1 bool `json:"day1"`
		Day2 bool `json:"day2"`
		Day3 bool `json:"day3"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	if err := db.Model(user).Updates(map[string]bool{
		"entry_day1": req.Day1,
		"entry_day2": req.Day2,
		"entry_day3": req.Day3,
	}).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

// PATCH /users/:id/password
func changeUserPassword(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	user := &User{}
	if err := db.First(user, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	req := struct {
		Password string `json:"password" validate:"printascii,required"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	salt := generateRandomString()
	encrypted := toHash(req.Password, salt)

	if err := db.Model(user).Updates(&User{Salt: salt, EncryptedPassword: encrypted}).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

// PATCH /users/:id/permission
func changeUserPermission(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	user := &User{}
	if err := db.First(user, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	req := struct {
		Permission int `json:"permission" validate:"min=0,max=2"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	if err := db.Model(user).Update("permission", req.Permission).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /users/:id/requests
func getUserRequests(c echo.Context) error {
	type resultRequestUser struct {
		ID        uint      `json:"id"`
		UserID    uint      `json:"user_id"`
		Num       int       `json:"num"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	type resultItem struct {
		ID       uint                 `json:"id"`
		Item     *Item                `json:"item"`
		Requests []*resultRequestUser `json:"requests"`
	}
	type userPriority struct {
		UserID    uint      `json:"user_id"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	type priority struct {
		Priority int             `json:"priority"`
		Users    []*userPriority `json:"users"`
	}
	type resultCircle struct {
		CircleID   uint          `json:"circle_id"`
		Priorities []*priority   `json:"priorities"`
		Items      []*resultItem `json:"items"`
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	var requests []*UserRequestItem
	if err := db.Preload("Item").Where(&UserRequestItem{UserID: uint(id)}).Find(&requests).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	var priorities []*UserCirclePriority
	if err := db.Where(&UserCirclePriority{UserID: uint(id)}).Find(&priorities).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	result := make([]*resultCircle, 0)
	circleMap := map[uint]*resultCircle{}
	itemMap := map[uint]*resultItem{}
	priorityMap := map[uint][]*priority{}

	for _, v := range priorities {
		if p, r := v.Priority1, 1; p != nil {
			if a, ok := priorityMap[*p]; ok {
				found := false
				for _, b := range a {
					if b.Priority == r {
						b.Users = append(b.Users, &userPriority{
							UserID:    v.UserID,
							UpdatedAt: v.UpdatedAt,
						})
						found = true
						break
					}
				}
				if !found {
					priorityMap[*p] = append(a, &priority{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}})
				}
			} else {
				priorityMap[*p] = []*priority{{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}}}
			}
		}
		if p, r := v.Priority2, 2; p != nil {
			if a, ok := priorityMap[*p]; ok {
				found := false
				for _, b := range a {
					if b.Priority == r {
						b.Users = append(b.Users, &userPriority{
							UserID:    v.UserID,
							UpdatedAt: v.UpdatedAt,
						})
						found = true
						break
					}
				}
				if !found {
					priorityMap[*p] = append(a, &priority{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}})
				}
			} else {
				priorityMap[*p] = []*priority{{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}}}
			}
		}
		if p, r := v.Priority3, 3; p != nil {
			if a, ok := priorityMap[*p]; ok {
				found := false
				for _, b := range a {
					if b.Priority == r {
						b.Users = append(b.Users, &userPriority{
							UserID:    v.UserID,
							UpdatedAt: v.UpdatedAt,
						})
						found = true
						break
					}
				}
				if !found {
					priorityMap[*p] = append(a, &priority{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}})
				}
			} else {
				priorityMap[*p] = []*priority{{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}}}
			}
		}
		if p, r := v.Priority4, 4; p != nil {
			if a, ok := priorityMap[*p]; ok {
				found := false
				for _, b := range a {
					if b.Priority == r {
						b.Users = append(b.Users, &userPriority{
							UserID:    v.UserID,
							UpdatedAt: v.UpdatedAt,
						})
						found = true
						break
					}
				}
				if !found {
					priorityMap[*p] = append(a, &priority{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}})
				}
			} else {
				priorityMap[*p] = []*priority{{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}}}
			}
		}
		if p, r := v.Priority5, 5; p != nil {
			if a, ok := priorityMap[*p]; ok {
				found := false
				for _, b := range a {
					if b.Priority == r {
						b.Users = append(b.Users, &userPriority{
							UserID:    v.UserID,
							UpdatedAt: v.UpdatedAt,
						})
						found = true
						break
					}
				}
				if !found {
					priorityMap[*p] = append(a, &priority{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}})
				}
			} else {
				priorityMap[*p] = []*priority{{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}}}
			}
		}
	}

	for _, v := range requests {
		c, ok := circleMap[v.Item.CircleID]
		if !ok {
			c = &resultCircle{
				CircleID: v.Item.CircleID,
			}
			if p, ok := priorityMap[v.Item.CircleID]; ok {
				sort.Slice(p, func(i, j int) bool {
					return p[i].Priority < p[j].Priority
				})
				c.Priorities = p
			} else {
				c.Priorities = make([]*priority, 0)
			}
			circleMap[v.Item.CircleID] = c
			result = append(result, c)
		}

		i, ok := itemMap[v.ItemID]
		if !ok {
			i = &resultItem{
				ID:   v.ItemID,
				Item: v.Item,
			}
			itemMap[v.ItemID] = i
			c.Items = append(c.Items, i)
		}

		u := &resultRequestUser{
			ID:        v.ID,
			UserID:    v.UserID,
			Num:       v.Num,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
		i.Requests = append(i.Requests, u)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CircleID < result[j].CircleID
	})

	return c.JSON(http.StatusOK, result)
}

// GET /users/:id/request-notes
func getUserRequestNotes(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	user := &User{}
	if err := db.First(user, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	ids := make([]uint, 0)
	if err := db.Model(&UserRequestNote{}).Where(&UserRequestNote{UserID: user.ID}).Order("created_at DESC").Pluck("id", &ids).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ids)
}

// GET /users/:id/circle-priorities
func getUserCirclePriorities(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	user := &User{}
	if err := db.First(user, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	var ucp []*UserCirclePriority
	if err := db.Where(&UserCirclePriority{UserID: user.ID}).Find(&ucp).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	res := &struct {
		Enterprise []uint `json:"enterprise"`
		Day1       []uint `json:"day1"`
		Day2       []uint `json:"day2"`
		Day3       []uint `json:"day3"`
	}{}
	res.Enterprise = make([]uint, 0)
	res.Day1 = make([]uint, 0)
	res.Day2 = make([]uint, 0)
	res.Day3 = make([]uint, 0)
	for _, v := range ucp {
		switch v.Day {
		case 0:
			res.Enterprise = v.GetPrioritySlice()
		case 1:
			res.Day1 = v.GetPrioritySlice()
		case 2:
			res.Day2 = v.GetPrioritySlice()
		case 3:
			res.Day3 = v.GetPrioritySlice()
		default:
			continue
		}
	}

	return c.JSON(http.StatusOK, res)
}

// GET /circles
// q?, days?
func getCircles(c echo.Context) error {
	days := convertStringsToUints(strings.Split(c.QueryParam("days"), ","))

	var arr []*Circle
	tx := db

	if len(days) > 0 {
		tx = tx.Where("day IN (?)", days)
	}
	query := "%" + c.QueryParam("q") + "%"
	tx = tx.Where("name LIKE ? OR author LIKE ?", query, query)

	if err := tx.Limit(100).Find(&arr).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, arr)
}

// GET /circles/requested
// day?
func getRequestedCircles(c echo.Context) error {
	req := struct {
		Day *uint `query:"day" validate:"min=0,max=3"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	tx := db.Joins("INNER JOIN (SELECT items.circle_id as id FROM items INNER JOIN user_request_items ON items.id = user_request_items.item_id GROUP BY items.circle_id) t ON circles.id = t.id")
	if req.Day != nil {
		tx = tx.Where("circles.day = ?", req.Day)
	}

	var arr []uint
	if err := tx.Model(&Circle{}).Pluck("circles.id", &arr).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, arr)
}

// GET /circles/:id
func getCircle(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	circle := Circle{}
	if err := db.First(&circle, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, &circle)
}

// GET /circles/:id/memos
func getCircleMemos(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	circle := Circle{}
	if err := db.First(&circle, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	var ids []uint
	if err := db.Model(&CircleMemo{}).Where("circle_id = ?", id).Pluck("id", &ids).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ids)
}

// POST /circles/:id/memos
func postCircleMemo(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	circle := Circle{}
	if err := db.First(&circle, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	req := struct {
		Content string `json:"content" validate:"required"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	memo := &CircleMemo{
		CircleID: circle.ID,
		UserID:   claim.GetUID(),
		Memo:     req.Content,
	}

	if err := db.Create(memo).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, map[string]uint{"id": memo.ID})
}

// GET /circles/:id/items
func getCircleItems(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	var items []*Item

	if err := db.Find(&items, &Item{CircleID: uint(id)}).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, items)
}

// GET /circle-memos/:mid
func getCircleMemo(c echo.Context) error {
	mid, err := strconv.Atoi(c.Param("mid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	memo := &CircleMemo{}
	if err := db.First(memo, mid).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, memo)
}

// DELETE /circle-memos/:mid
func deleteCircleMemo(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)
	mid, err := strconv.Atoi(c.Param("mid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	memo := &CircleMemo{}
	if err := db.First(memo, mid).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if memo.UserID != claim.GetUID() {
		return echo.NewHTTPError(http.StatusForbidden)
	}

	if err := db.Delete(memo).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

// POST /items
func postItem(c echo.Context) error {
	req := struct {
		CircleID uint   `json:"circle_id" validate:"required"`
		Name     string `json:"name"      validate:"required,max=100"`
		Price    int    `json:"price"     validate:"min=-1"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	circle := Circle{}
	if err := db.First(&circle, req.CircleID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "存在しないサークルIDです")
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	item := &Item{
		CircleID: req.CircleID,
		Name:     req.Name,
		Price:    req.Price,
	}
	if err := db.Create(item).Error; err != nil {
		if isMySQLDuplicatedRecordErr(err) {
			return echo.NewHTTPError(http.StatusConflict, "既に登録されている商品名です")
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, map[string]uint{"id": item.ID})
}

// GET /items/:id
func getItem(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	item := Item{}
	if err := db.First(&item, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error()
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, &item)
}

// PATCH /items/:id/name
func updateItemName(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	item := Item{}
	if err := db.First(&item, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error()
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	req := struct {
		Name int `json:"name"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	if err := db.Model(&item).Update("name", req.Name).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

// PATCH /items/:id/price
func updateItemPrice(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	item := Item{}
	if err := db.First(&item, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error()
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	req := struct {
		Price int `json:"price"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	if err := db.Model(&item).Update("price", req.Price).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /requests
func getRequests(c echo.Context) error {
	type resultRequestUser struct {
		ID        uint      `json:"id"`
		UserID    uint      `json:"user_id"`
		Num       int       `json:"num"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	type resultItem struct {
		ID       uint                 `json:"id"`
		Item     *Item                `json:"item"`
		Requests []*resultRequestUser `json:"requests"`
	}
	type userPriority struct {
		UserID    uint      `json:"user_id"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	type priority struct {
		Priority int             `json:"priority"`
		Users    []*userPriority `json:"users"`
	}
	type resultCircle struct {
		CircleID   uint          `json:"circle_id"`
		Priorities []*priority   `json:"priorities"`
		Items      []*resultItem `json:"items"`
	}

	var requests []*UserRequestItem
	if err := db.Preload("Item").Find(&requests).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	var priorities []*UserCirclePriority
	if err := db.Find(&priorities).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	var result []*resultCircle
	circleMap := map[uint]*resultCircle{}
	itemMap := map[uint]*resultItem{}
	priorityMap := map[uint][]*priority{}

	for _, v := range priorities {
		if p, r := v.Priority1, 1; p != nil {
			if a, ok := priorityMap[*p]; ok {
				found := false
				for _, b := range a {
					if b.Priority == r {
						b.Users = append(b.Users, &userPriority{
							UserID:    v.UserID,
							UpdatedAt: v.UpdatedAt,
						})
						found = true
						break
					}
				}
				if !found {
					priorityMap[*p] = append(a, &priority{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}})
				}
			} else {
				priorityMap[*p] = []*priority{{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}}}
			}
		}
		if p, r := v.Priority2, 2; p != nil {
			if a, ok := priorityMap[*p]; ok {
				found := false
				for _, b := range a {
					if b.Priority == r {
						b.Users = append(b.Users, &userPriority{
							UserID:    v.UserID,
							UpdatedAt: v.UpdatedAt,
						})
						found = true
						break
					}
				}
				if !found {
					priorityMap[*p] = append(a, &priority{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}})
				}
			} else {
				priorityMap[*p] = []*priority{{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}}}
			}
		}
		if p, r := v.Priority3, 3; p != nil {
			if a, ok := priorityMap[*p]; ok {
				found := false
				for _, b := range a {
					if b.Priority == r {
						b.Users = append(b.Users, &userPriority{
							UserID:    v.UserID,
							UpdatedAt: v.UpdatedAt,
						})
						found = true
						break
					}
				}
				if !found {
					priorityMap[*p] = append(a, &priority{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}})
				}
			} else {
				priorityMap[*p] = []*priority{{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}}}
			}
		}
		if p, r := v.Priority4, 4; p != nil {
			if a, ok := priorityMap[*p]; ok {
				found := false
				for _, b := range a {
					if b.Priority == r {
						b.Users = append(b.Users, &userPriority{
							UserID:    v.UserID,
							UpdatedAt: v.UpdatedAt,
						})
						found = true
						break
					}
				}
				if !found {
					priorityMap[*p] = append(a, &priority{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}})
				}
			} else {
				priorityMap[*p] = []*priority{{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}}}
			}
		}
		if p, r := v.Priority5, 5; p != nil {
			if a, ok := priorityMap[*p]; ok {
				found := false
				for _, b := range a {
					if b.Priority == r {
						b.Users = append(b.Users, &userPriority{
							UserID:    v.UserID,
							UpdatedAt: v.UpdatedAt,
						})
						found = true
						break
					}
				}
				if !found {
					priorityMap[*p] = append(a, &priority{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}})
				}
			} else {
				priorityMap[*p] = []*priority{{Priority: r, Users: []*userPriority{{UserID: v.UserID, UpdatedAt: v.UpdatedAt}}}}
			}
		}
	}

	for _, v := range requests {
		c, ok := circleMap[v.Item.CircleID]
		if !ok {
			c = &resultCircle{
				CircleID: v.Item.CircleID,
			}
			if p, ok := priorityMap[v.Item.CircleID]; ok {
				sort.Slice(p, func(i, j int) bool {
					return p[i].Priority < p[j].Priority
				})
				c.Priorities = p
			} else {
				c.Priorities = make([]*priority, 0)
			}
			circleMap[v.Item.CircleID] = c
			result = append(result, c)
		}

		i, ok := itemMap[v.ItemID]
		if !ok {
			i = &resultItem{
				ID:   v.ItemID,
				Item: v.Item,
			}
			itemMap[v.ItemID] = i
			c.Items = append(c.Items, i)
		}

		u := &resultRequestUser{
			ID:        v.ID,
			UserID:    v.UserID,
			Num:       v.Num,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
		i.Requests = append(i.Requests, u)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CircleID < result[j].CircleID
	})

	return c.JSON(http.StatusOK, result)
}

// POST /requests
func postRequest(c echo.Context) error {
	req := struct {
		UserID uint `json:"user_id" validate:"required"`
		ItemID uint `json:"item_id" validate:"required"`
		Num    int  `json:"num"     validate:"min=1"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	user := &User{}
	if err := db.First(user, req.UserID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "the user is not found")
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	circle := Circle{}
	if err := db.Where("id = (?)", db.Model(Item{ID: req.ItemID}).Select("circle_id").QueryExpr()).Take(&circle).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "the item is not found")
		}
		c.Logger().Error()
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	ur := &UserRequestItem{
		UserID: req.UserID,
		ItemID: req.ItemID,
		Num:    req.Num,
	}

	err := db.
		Set("gorm:insert_option", fmt.Sprintf("ON DUPLICATE KEY UPDATE num = %d, updated_at = current_timestamp(6)", ur.Num)).
		Create(&ur).
		Error
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, map[string]uint{"id": ur.ID})
}

// GET /requests/:id
func getRequest(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	ur := &UserRequestItem{}
	if err := db.Preload("Item").Take(ur, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if claim.GetUID() != ur.UserID && getPermission(c) <= LevelUser {
		return echo.NewHTTPError(http.StatusForbidden)
	}

	return c.JSON(http.StatusOK, ur)
}

// PATCH /requests/:id
func editRequest(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "対象のリクエストは存在しません")
	}

	req := struct {
		Num int `json:"num" validate:"min=1"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	ur := &UserRequestItem{}
	if err := db.Take(ur, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound, "対象のリクエストは存在しません")
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if claim.GetUID() != ur.UserID && getPermission(c) <= LevelUser {
		return echo.NewHTTPError(http.StatusForbidden)
	}

	circle := Circle{}
	if err := db.Where("id = (?)", db.Model(Item{ID: ur.ItemID}).Select("circle_id").QueryExpr()).Take(&circle).Error; err != nil {
		c.Logger().Error()
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if dl, ok := getDayDeadLine(circle.Day); ok && dl.Before(time.Now()) && getPermission(c) <= LevelUser {
		return echo.NewHTTPError(http.StatusForbidden, "既にリクエスト締め切りが過ぎています")
	}

	if err := db.Model(ur).Updates(&UserRequestItem{Num: req.Num}).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

// DELETE /requests/:id
func deleteRequest(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "対象のリクエストは存在しません")
	}

	ur := &UserRequestItem{}
	if err := db.Take(ur, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound, "対象のリクエストは存在しません")
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if claim.GetUID() != ur.UserID && getPermission(c) <= LevelUser {
		return echo.NewHTTPError(http.StatusForbidden)
	}

	circle := Circle{}
	if err := db.Where("id = (?)", db.Model(Item{ID: ur.ItemID}).Select("circle_id").QueryExpr()).Take(&circle).Error; err != nil {
		c.Logger().Error()
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if dl, ok := getDayDeadLine(circle.Day); ok && dl.Before(time.Now()) && getPermission(c) <= LevelUser {
		return echo.NewHTTPError(http.StatusForbidden, "既にリクエスト締め切りが過ぎています")
	}

	if err := db.Delete(ur).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /request-notes
func getRequestNotes(c echo.Context) error {
	ids := make([]uint, 0)
	if err := db.Model(&UserRequestNote{}).Order("created_at DESC").Pluck("id", &ids).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ids)
}

// POST /request-notes
func postRequestNote(c echo.Context) error {
	return postMyRequestNotes(c)
}

// GET /request-notes/:id
func getRequestNote(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	urn := &UserRequestNote{}
	if err := db.Take(urn, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if claim.GetUID() != urn.UserID && getPermission(c) <= LevelUser {
		return echo.NewHTTPError(http.StatusForbidden)
	}

	return c.JSON(http.StatusOK, urn)
}

// DELETE /request-notes/:id
func deleteRequestNote(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	urn := &UserRequestNote{}
	if err := db.Take(urn, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if claim.GetUID() != urn.UserID {
		return echo.NewHTTPError(http.StatusForbidden)
	}

	if err := db.Delete(urn).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

/*
// POST /assignment-group-notes
func postAssignmentNote(c echo.Context) error {
	claim := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)

	req := struct {
		GroupID uint   `json:"group_id" validate:"required,min=1"`
		Content string `json:"content" validate:"required"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	agn := &AssignmentGroupNote{
		WriterID: claim.GetUID(),
		GroupID:  req.GroupID,
		Content:  req.Content,
	}
	if err := db.Create(agn).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{"id": agn.ID})
}

// GET /assignment-group-notes/:id
func getAssignmentNote(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	agn := &AssignmentGroupNote{}
	if err := db.Take(agn, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, agn)
}

// DELETE /assignment-group-notes/:id
func deleteAssignmentNote(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	agn := &AssignmentGroupNote{}
	if err := db.Take(agn, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := db.Delete(agn).Error; err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}
*/

// GET /deadlines
func getDeadLines(c echo.Context) error {
	res := struct {
		Enterprise *time.Time `json:"enterprise"`
		Day1       *time.Time `json:"day1"`
		Day2       *time.Time `json:"day2"`
		Day3       *time.Time `json:"day3"`
	}{}

	if t, ok := getDayDeadLine(0); ok {
		res.Enterprise = &t
	}
	if t, ok := getDayDeadLine(1); ok {
		res.Day1 = &t
	}
	if t, ok := getDayDeadLine(2); ok {
		res.Day2 = &t
	}
	if t, ok := getDayDeadLine(3); ok {
		res.Day3 = &t
	}

	return c.JSON(http.StatusOK, res)
}

// PUT /deadlines
func setDeadLines(c echo.Context) error {
	req := struct {
		Day      int       `json:"day"      validate:"min=0,max=3"`
		Datetime time.Time `json:"datetime" validate:"required"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	if err := setDayDeadLine(req.Day, req.Datetime); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}
