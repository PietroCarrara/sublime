package legendastv

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/PietroCarrara/sublime/pkg/guessit"
	"github.com/PuerkitoBio/goquery"
)

type mediaType string
type subType string

const (
	mediaTypeMovie mediaType = "M"
	mediaTypeShow  mediaType = "S"

	subTypeAny       subType = "-"
	subTypeHighlight subType = "d"
	subTypePack      subType = "p"
)

type mediaEntry struct {
	ID     int
	Title  string
	Season int
	Type   mediaType
}

type subtitleEntry struct {
	ID    string
	Title string
	Type  subType
}

func getEntries(client *http.Client, info guessit.Information) ([]*mediaEntry, error) {
	res := []*mediaEntry{}

	title := url.PathEscape(info.Title)

	r, err := client.Get(fmt.Sprintf("http://legendas.tv/legenda/sugestao/%s", title))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	obj, err := gabs.ParseJSONBuffer(r.Body)
	if err != nil {
		return nil, err
	}

	for _, obj := range obj.Children() {
		data := map[string]string{
			"id_filme":    "",
			"dsc_nome":    "",
			"dsc_nome_br": "",
			"tipo":        "",
			"temporada":   "",
		}

		for field := range data {
			value, ok := obj.Path("_source." + field).Data().(string)
			if !ok {
				continue
			}
			data[field] = value
		}

		// TODO: More forgiving way to match the title
		if strings.ToLower(data["dsc_nome"]) == strings.ToLower(info.Title) || strings.ToLower(data["dsc_nome_br"]) == strings.ToLower(info.Title) {
			entry, err := entryFromFields(data["id_filme"], data["dsc_nome"], data["temporada"], data["tipo"])
			if err != nil {
				log.Printf("legendastv: %s", err)
				continue
			}

			if entry.Season == info.Season {
				res = append(res, entry)
			}
		}
	}

	if len(res) > 0 {
		return res, nil
	}

	if info.Season != 0 {
		err = fmt.Errorf(`could not find title "%s", season %d`, info.Title, info.Season)
	} else {
		err = fmt.Errorf(`could not find title "%s"`, info.Title)
	}
	return nil, err
}

func entryFromFields(id, title, season, typ string) (*mediaEntry, error) {
	ID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	Season := 0
	if season != "" {
		Season, err = strconv.Atoi(season)
		if err != nil {
			return nil, err
		}
	}

	var Type mediaType
	switch typ {
	case "S":
		Type = mediaTypeShow
	case "M":
		Type = mediaTypeMovie
	default:
		return nil, fmt.Errorf(`could not find media type "%s"`, typ)
	}

	return &mediaEntry{
		ID:     ID,
		Season: Season,
		Title:  title,
		Type:   Type,
	}, nil
}

var subIDRegex = regexp.MustCompile(`(?i)/download/([\w\d]+)/`)

func (e mediaEntry) ListSubtitles(client *http.Client, typ subType, languageID int) ([]subtitleEntry, error) {
	page := 0
	res := make([]subtitleEntry, 0)

	for {
		url := fmt.Sprintf(
			"http://legendas.tv/legenda/busca/-/%d/%s/%d/%d",
			languageID,
			typ,
			page,
			e.ID,
		)

		r, err := client.Get(url)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()

		doc, err := goquery.NewDocumentFromReader(r.Body)
		if err != nil {
			return nil, err
		}

		doc.Find(".gallery > article > div").Each(func(i int, element *goquery.Selection) {
			title := element.Find("p:not(.data) > a").Text()
			href, ok := element.Find("p:not(.data) > a").Attr("href")
			if !ok {
				log.Printf(`legendastv: couldn't find subtitle download link for subtitle "%s"`, title)
				return
			}

			submatches := subIDRegex.FindStringSubmatch(href)
			if len(submatches) < 2 {
				log.Printf(`legendastv: couldn't find subtitle download link for subtitle "%s"`, title)
			}
			id := submatches[1]

			typ := subTypeAny
			if element.HasClass("pack") {
				typ = subTypePack
			} else if element.HasClass("destaque") {
				typ = subTypeHighlight
			}

			res = append(res, subtitleEntry{
				ID:    id,
				Title: title,
				Type:  typ,
			})
		})

		// TODO: Check for the next page
		break
	}

	return res, nil
}

func (s subtitleEntry) DownloadContents(c *http.Client) (io.ReadCloser, error) {
	url := fmt.Sprintf("http://legendas.tv/downloadarquivo/%s", s.ID)

	r, err := c.Get(url)
	r.Body.Close()
	if err != nil {
		return nil, err
	}

	location, err := r.Location()
	if err != nil {
		return nil, err
	}

	r, err = c.Get(location.String())
	if err != nil {
		r.Body.Close()
		return nil, err
	}

	return r.Body, nil
}
