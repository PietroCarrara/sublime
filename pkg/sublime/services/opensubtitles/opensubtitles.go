package opensubtitles

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/PietroCarrara/sublime/pkg/guessit"
	"github.com/PietroCarrara/sublime/pkg/sublime"
	"github.com/oz/osdb"
	"golang.org/x/text/language"
)

const name = "opensubtitles"

var langToISO639 = map[language.Tag][]string{
	language.BrazilianPortuguese: {"pob", "pb"},
}

type OpenSubtitles struct {
	c        *osdb.Client
	username string
	password string
}

type OpenSubtitlesSubtitle struct {
	s osdb.Subtitle
	t *sublime.FileTarget
	c *osdb.Client
}

func init() {
	o := &OpenSubtitles{}

	sublime.Services[o.GetName()] = o
}

func (o *OpenSubtitles) GetName() string {
	return name
}

func (o *OpenSubtitles) GetCandidatesForFiles(files []*sublime.FileTarget, langs []language.Tag) <-chan sublime.SubtitleCandidate {
	langList := make([]string, len(langs))
	for i, lang := range langs {
		if iso639, ok := langToISO639[lang]; ok {
			langList[i] = iso639[0]
		} else {
			l, _ := lang.Base()
			langList[i] = l.ISO3()
		}
	}
	langsString := strings.Join(langList, ",")

	channel := make(chan sublime.SubtitleCandidate)
	go func() {
		// Loop over every file
		for _, file := range files {
			args := map[string]string{
				"query":         file.GetName(),
				"sublanguageid": langsString,
			}
			params := []interface{}{
				o.c.Token,
				[]map[string]string{args},
			}
			res, err := o.c.SearchSubtitles(&params)
			if err != nil {
				log.Printf("opensubtitles: %s\n", err)
				// Go to the next file
				continue
			}

			for _, sub := range res {
				// Check if seasons match
				if season := file.GetInfo().Season; season != 0 {
					if sub.SeriesSeason != strconv.Itoa(season) {
						continue
					}
				}

				candidate := OpenSubtitlesSubtitle{
					s: sub,
					t: file,
					c: o.c,
				}
				channel <- candidate
			}
		}
		close(channel)
	}()

	return channel
}

func (o *OpenSubtitles) SetConfig(name, value string) error {
	switch name {
	case "username":
		o.username = value
	case "password":
		o.password = value
	default:
		return fmt.Errorf(`option "%s" was not found`, name)
	}

	return nil
}

func (o *OpenSubtitles) Initialize() error {
	c, err := osdb.NewClient()
	if err != nil {
		return err
	}
	o.c = c

	return o.c.LogIn(o.username, o.password, "")
}

func (s OpenSubtitlesSubtitle) GetFormatExtension() string {
	// We always download srt subtitles
	return "srt"
}

func (s OpenSubtitlesSubtitle) GetService() string {
	return name
}

func (s OpenSubtitlesSubtitle) GetRanking() float32 {
	cnt, err := strconv.Atoi(s.s.SubDownloadsCnt)
	if err != nil {
		return 0
	}
	return float32(cnt)
}

func (s OpenSubtitlesSubtitle) GetFileTarget() *sublime.FileTarget {
	return s.t
}

func (s OpenSubtitlesSubtitle) GetLang() language.Tag {
	for lang, values := range langToISO639 {
		for _, iso639 := range values {
			if iso639 == s.s.ISO639 {
				return lang
			}
		}
	}

	return language.Make(s.s.ISO639)
}

func (s OpenSubtitlesSubtitle) GetInfo() guessit.Information {
	return guessit.Parse(s.s.SubFileName)
}

func (s OpenSubtitlesSubtitle) Open() (io.ReadCloser, error) {
	// TODO: Rework the library so we don't have to do this disk operation
	tmpFile, err := ioutil.TempFile(os.TempDir(), "sublime-opensubtitles-")
	if err != nil {
		return nil, err
	}
	filename := tmpFile.Name()
	tmpFile.Close()

	err = s.c.DownloadTo(&s.s, filename)
	if err != nil {
		return nil, err
	}

	return os.Open(filename)
}
