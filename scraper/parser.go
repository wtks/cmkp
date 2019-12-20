package main

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"io"
	"strconv"
	"strings"
)

func parseBoothSyllabary(r io.Reader) (map[string]string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := map[string]string{}
	doc.Find("div.md-circlelist > div > ul > li > div > a").Each(func(i int, s *goquery.Selection) {
		url, has := s.Attr("href")
		if has {
			result[strings.TrimSpace(s.Find("div").Text())] = strings.TrimPrefix(url, "/Booth/")
		}
	})
	return result, nil
}

func parseBooth(r io.Reader) (*Booth, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	booth := &Booth{}
	doc.Find("table.m-companytable__table tr").Has("th").Each(func(i int, tr *goquery.Selection) {
		td := tr.Find("td")
		switch i {
		case 0:
			booth.Name = strings.TrimSpace(td.Text())
		case 1:
			booth.No, _ = strconv.Atoi(strings.TrimSpace(td.Text()))
		case 2:
			booth.Summery = strings.TrimSpace(td.Text())
		case 3:
			booth.Description = strings.TrimSpace(td.Text())
		case 4:
			td.Find("a").Each(func(_ int, a *goquery.Selection) {
				url, has := a.Attr("href")
				if has {
					booth.Links = append(booth.Links, url)
				}
			})
		}
	})
	return booth, nil
}

func extractModel(r io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return doc.Find("#TheModel").Text(), nil
}

func parseCircleList(modelJson string) ([]Circle, error) {
	var m struct {
		Circles []Circle `json:"Circles"`
	}
	if err := json.Unmarshal([]byte(modelJson), &m); err != nil {
		return nil, errors.WithStack(err)
	}
	return m.Circles, nil
}
