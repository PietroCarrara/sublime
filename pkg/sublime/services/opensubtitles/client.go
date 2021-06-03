package opensubtitles

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client struct {
	key     string
	session *http.Client
	baseUrl *url.URL
}

// SubtitleArgs represents the arguments to the subtitles API call
// Documentation: https://opensubtitles.stoplight.io/docs/opensubtitles-api/open_api.json/paths/~1api~1v1~1subtitles/get
type SubtitlesArgs struct {
	ID                string `json:"id,omitempty"`
	Page              int    `json:"page,omitempty"`
	ImdbID            string `json:"imdb_id,omitempty"`
	TmdbID            int    `json:"tmdb_id,omitempty"`
	Languages         string `json:"languages,omitempty"`
	Type              string `json:"type,omitempty"`
	MovieHash         string `json:"movie_hash,omitempty"`
	UserID            string `json:"user_id,omitempty"`
	HearingImpaired   string `json:"hearing_impaired,omitempty"`
	ForeignPartsOnly  string `json:"foreign_parts_only,omitempty"`
	TrustedSources    string `json:"trusted_sources,omitempty"`
	MachineTranslated string `json:"machine_translated,omitempty"`
	AiTranslated      string `json:"ai_translated,omitempty"`
	OrderBy           string `json:"order_by,omitempty"`
	OrderDirection    string `json:"order_direction,omitempty"`
	ParentFeatureID   int    `json:"parent_feature_id,omitempty"`
	ParentImdbID      int    `json:"parent_imdb_id,omitempty"`
	ParentTmdbID      int    `json:"parent_tmdb_id,omitempty"`
	Query             string `json:"query,omitempty"`
	SeasonNumber      int    `json:"season_number,omitempty"`
	EpisodeNumber     int    `json:"episode_number,omitempty"`
	Year              int    `json:"year,omitempty"`
}

type SubtitlesResult struct {
	TotalPages int        `json:"total_pages,omitempty"`
	TotalCount int        `json:"total_count,omitempty"`
	Page       string     `json:"page,omitempty"`
	Data       []Subtitle `json:"data,omitempty"`
}

type Subtitle struct {
	ID         string
	Type       string
	Attributes struct {
		Language          string  `json:"language,omitempty"`
		DownloadCount     int     `json:"download_count,omitempty"`
		NewDownloadCount  int     `json:"new_download_count,omitempty"`
		HearingImpaired   bool    `json:"hearing_impaired,omitempty"`
		HD                bool    `json:"hd,omitempty"`
		Format            string  `json:"format,omitempty"`
		FPS               float32 `json:"fps,omitempty"`
		Votes             int     `json:"votes,omitempty"`
		Points            int     `json:"points,omitempty"`
		Ratings           float32 `json:"ratings,omitempty"`
		FromTrusted       bool    `json:"from_trusted,omitempty"`
		ForeignPartsOnly  bool    `json:"foreign_parts_only,omitempty"`
		AutoTranslation   bool    `json:"auto_translation,omitempty"`
		AiTranslated      bool    `json:"ai_translated,omitempty"`
		MachineTranslated bool    `json:"machine_translated,omitempty"`
		UploadDate        string  `json:"upload_date,omitempty"`
		Release           string  `json:"release,omitempty"`
		Comments          string  `json:"comments,omitempty"`
		LegacySubtitleId  int     `json:"legacy_subtitle_id,omitempty"`
		Uploader          struct {
			UploaderID int    `json:"uploader_id,omitempty"`
			Name       string `json:"name,omitempty"`
			Rank       string `json:"rank,omitempty"`
		} `json:"uploader,omitempty"`
		FeatureDetails struct {
			FeatureID   int    `json:"feature_id,omitempty"`
			FeatureType string `json:"feature_type,omitempty"`
			Year        int    `json:"year,omitempty"`
			Title       string `json:"title,omitempty"`
			MovieName   string `json:"movie_name,omitempty"`
			ImdbID      int    `json:"imdb_id,omitempty"`
			TmdbID      int    `json:"tmdb_id,omitempty"`
		} `json:"feature_details,omitempty"`
		Url string `json:"url,omitempty"`
		// related_links can be a single object ( {} ) or multiple objects ( [{}, ..., {}] )
		// RelatedLinks struct {
		// 	Label  string `json:"label,omitempty"`
		// 	Url    string `json:"url,omitempty"`
		// 	ImgUrl string `json:"img_url,omitempty"`
		// } `json:"related_links,omitempty"`
		Files []struct {
			FileID   int    `json:"file_id,omitempty"`
			CdNumber int    `json:"cd_number,omitempty"`
			FileName string `json:"file_name,omitempty"`
		} `json:"files,omitempty"`
	} `json:"attributes,omitempty"`
}

type DownloadResult struct {
	Link string `json:"link,omitempty"`
}

func NewClient(key string) *Client {
	url, _ := url.Parse("https://api.opensubtitles.com")

	return &Client{
		session: http.DefaultClient,
		baseUrl: url,
		key:     key,
	}
}

func (c *Client) GetSubtitles(args SubtitlesArgs) (*SubtitlesResult, error) {
	// Get the supplied arguments
	params, err := objToUrl(args)
	if err != nil {
		return nil, err
	}
	obj, err := c.ApiCall("/api/v1/subtitles?"+params, "GET", []byte{})
	if err != nil {
		return nil, err
	}

	var res SubtitlesResult
	err = json.Unmarshal(obj, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) GetDownload(sub Subtitle) (io.ReadCloser, error) {
	byts, err := c.ApiCall(
		fmt.Sprintf("/api/v1/download?file_id=%s&sub_format=srt", sub.ID),
		"POST",
		[]byte{},
	)
	if err != nil {
		return nil, err
	}

	var link DownloadResult
	err = json.Unmarshal(byts, &link)
	if err != nil {
		return nil, err
	}

	res, err := c.session.Get(link.Link)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (c *Client) ApiCall(route, method string, body []byte) ([]byte, error) {
	empty := []byte{}

	url, err := url.Parse(route)
	if err != nil {
		return empty, err
	}

	req, err := http.NewRequest(
		method,
		c.baseUrl.ResolveReference(url).String(),
		bytes.NewReader(body),
	)
	if err != nil {
		return empty, nil
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Api-Key", c.key)
	req.Header.Add("Accept", "*/*")

	res, err := c.session.Do(req)
	if err != nil {
		return empty, err
	}
	defer res.Body.Close()

	byts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return empty, nil
	}

	return byts, err
}

func objToUrl(v interface{}) (string, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return "nil", err
	}
	var params map[string]interface{}
	err = json.Unmarshal(body, &params)
	if err != nil {
		return "nil", err
	}

	vals := url.Values{}
	for key, val := range params {
		vals.Add(key, fmt.Sprint(val))
	}

	return vals.Encode(), nil
}
