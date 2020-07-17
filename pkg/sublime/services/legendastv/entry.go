package legendastv

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/PietroCarrara/sublime/pkg/guessit"
)

type mediaType int

const (
	mediaTypeMovie mediaType = iota
	mediaTypeShow
)

type entry struct {
	ID     int
	Title  string
	Season int
	Type   mediaType
}

func getEntry(client *http.Client, info guessit.Information) (*entry, error) {
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
			"id_filme":  "",
			"dsc_nome":  "",
			"tipo":      "",
			"temporada": "",
		}

		for field := range data {
			value, ok := obj.Path("_source." + field).Data().(string)
			if !ok {
				continue
			}
			data[field] = value
		}

		// TODO: More forgiving way to match the title
		if strings.ToLower(data["dsc_nome"]) == strings.ToLower(info.Title) {
			entry, err := (entryFromFields(data["id_filme"], data["dsc_nome"], data["temporada"], data["tipo"]))
			if err != nil {
				return nil, err
			}

			if entry.Season == info.Season {
				return entry, nil
			}
		}
	}

	if info.Season != 0 {
		err = fmt.Errorf(`could not find title "%s", season %d`, info.Title, info.Season)
	} else {
		err = fmt.Errorf(`could not find title "%s"`, info.Title)
	}

	return nil, err
}

func entryFromFields(id, title, season, typ string) (*entry, error) {
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

	return &entry{
		ID:     ID,
		Season: Season,
		Title:  title,
		Type:   Type,
	}, nil
}
