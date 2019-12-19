package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/wtks/cmkp/backend/model"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

func newValidator() *Validator {
	v := validator.New()
	return &Validator{validator: v}
}

func mustParseInt(str string) int {
	n, _ := strconv.Atoi(str)
	return n
}

func getUserId(ctx context.Context) int {
	return ctx.Value("userId").(int)
}

func getUserRole(ctx context.Context) model.Role {
	return ctx.Value("role").(model.Role)
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
