package opensubtitles

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/PietroCarrara/sublime/pkg/guessit"
	"github.com/PietroCarrara/sublime/pkg/sublime"
	"golang.org/x/text/language"
)

type OpenSubtitles struct {
	c        *Client
	key      string
	username string
	password string
}

type OpenSubtitlesSubtitle struct {
	s Subtitle
	t *sublime.FileTarget
	c *Client
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
		langsString[i] = lang.String()
	}
	langList := strings.Join(langsString, ",")

	channel := make(chan sublime.SubtitleCandidate)
	go func() {
		// Loop over every file
		for _, file := range files {
			page := 1
			// Loop over every page
			for {
				res, err := o.c.GetSubtitles(SubtitlesArgs{
					Query:     file.GetName(),
					Languages: langList,
					Page:      page,
				})
				if err != nil {
					log.Printf("opensubtitles: %s\n", err)
					break
				}

				for _, sub := range res.Data {
					candidate := OpenSubtitlesSubtitle{
						s: sub,
						t: file,
						c: o.c,
					}
					channel <- candidate
				}

				if page >= res.TotalPages {
					break
				}
				page++
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
	o.c = NewClient(o.key)

	// TODO: Login

	return nil
}

func (s OpenSubtitlesSubtitle) GetFormatExtension() string {
	// We always download srt subtitles
	return "srt"
}

func (s OpenSubtitlesSubtitle) GetFileTarget() *sublime.FileTarget {
	return s.t
}

func (s OpenSubtitlesSubtitle) GetLang() language.Tag {
	return language.Make(s.s.Attributes.Language)
}

func (s OpenSubtitlesSubtitle) GetInfo() guessit.Information {
	return guessit.Parse(s.s.Attributes.Release)
}

func (s OpenSubtitlesSubtitle) Open() (io.ReadCloser, error) {
	return s.c.GetDownload(s.s)
}
