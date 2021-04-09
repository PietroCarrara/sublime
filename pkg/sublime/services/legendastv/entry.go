package legendastv

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/PietroCarrara/sublime/pkg/guessit"
	"github.com/PuerkitoBio/goquery"
	"github.com/agnivade/levenshtein"
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

// getEntries returns a list of possible candidates matching the given information
// The first entries are likely to be the correct entry, while the last ones are not
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

	l := strings.ToLower

	for _, obj := range obj.Children() {
		var release int = 0
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
		if value, ok := obj.Path("_source.dsc_data_lancamento").Data().(string); ok {
			if date, err := strconv.Atoi(value); err == nil {
				release = date
			}
		}
		name := data["dsc_nome"]
		if levenshtein.ComputeDistance(l(data["dsc_nome_br"]), l(info.Title)) < levenshtein.ComputeDistance(l(name), l(info.Title)) {
			name = data["dsc_nome_br"]
		}

		entry, err := entryFromFields(data["id_filme"], name, data["temporada"], data["tipo"])
		if err != nil {
			log.Printf("legendastv: %s", err)
			continue
		}

		if info.Season == 0 && entry.Season == 0 {
			// If we're searching
			if info.Year > 0 && release > 0 && math.Abs(float64(info.Year-release)) <= 1 {
				res = append(res, entry)
			}
		} else {
			// If we're searching for a show, only add matching seasons
			if entry.Season == info.Season {
				res = append(res, entry)
			}
		}
	}

	if len(res) > 0 {
		// Sort by name proximity
		sort.Slice(res, func(a, b int) bool {
			dA := levenshtein.ComputeDistance(l(res[a].Title), l(info.Title))
			dB := levenshtein.ComputeDistance(l(res[b].Title), l(info.Title))

			return dA < dB
		})

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

		// Check if there is a next page
		if len(doc.Find(".load_more").Nodes) == 0 {
			break
		}

		page++
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
