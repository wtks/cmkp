package main

import (
	"github.com/pkg/errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type client struct {
	http.Client
}

func makeClient() *client {
	jar, _ := cookiejar.New(nil)
	return &client{
		Client: http.Client{Jar: jar},
	}
}

func (c *client) login(username, password string) error {
	data := url.Values{}
	data.Set("state", "/")
	data.Set("ReturnUrl", "https://webcatalog.circle.ms/Account/Login")
	data.Set("Username", username)
	data.Set("password", password)

	resp, err := c.PostForm("https://auth2.circle.ms/", data)
	if err != nil {
		return errors.WithStack(err)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("Login failed")
	}
	return nil
}

func (c *client) getCircleList(query url.Values) (*http.Response, error) {
	u, _ := url.Parse("https://webcatalog.circle.ms/Circle/List")
	u.RawQuery = query.Encode()
	return c.Get(u.String())
}

func (c *client) getBoothSyllabary() (*http.Response, error) {
	return c.Get("https://webcatalog.circle.ms/Booth/Syllabary")
}

func (c *client) getBooth(boothId string) (*http.Response, error) {
	return c.Get("https://webcatalog.circle.ms/Booth/" + boothId)
}
