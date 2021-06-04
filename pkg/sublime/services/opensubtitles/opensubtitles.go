package opensubtitles

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/PietroCarrara/sublime/pkg/guessit"
	"github.com/PietroCarrara/sublime/pkg/sublime"
	"github.com/oz/osdb"
	"golang.org/x/text/language"
)

var langToISO639 = map[language.Tag]string{
	language.BrazilianPortuguese: "pob",
}

type OpenSubtitles struct {
	c        *osdb.Client
	key      string
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
	return "opensubtitles"
}

func (o *OpenSubtitles) GetCandidatesForFiles(files []*sublime.FileTarget, langs []language.Tag) <-chan sublime.SubtitleCandidate {
	langsString := make([]string, len(langs))
	for i, lang := range langs {
		if iso639, ok := langToISO639[lang]; ok {
			langsString[i] = iso639
		} else {
			l, _ := lang.Base()
			langsString[i] = l.ISO3()
		}
	}
	langList := strings.Join(langsString, ",")

	channel := make(chan sublime.SubtitleCandidate)
	go func() {
		// Loop over every file
		for _, file := range files {
			args := map[string]string{
				"query":         file.GetName(),
				"sublanguageid": langList,
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
	case "key":
		o.key = value
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

func (s OpenSubtitlesSubtitle) GetFileTarget() *sublime.FileTarget {
	return s.t
}

func (s OpenSubtitlesSubtitle) GetLang() language.Tag {
	for lang, iso639 := range langToISO639 {
		if iso639 == s.s.ISO639 {
			return lang
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
