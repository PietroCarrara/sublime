package legendastv

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/PietroCarrara/sublime/pkg/guessit"
	"github.com/PietroCarrara/sublime/pkg/sublime"
	"github.com/mholt/archiver/v3"
	"golang.org/x/text/language"
)

var (
	rar15MagicNumber = []byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x00}
	rar50MagicNumber = []byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x01, 0x00}
	zipMagicNumber   = []byte{0x50, 0x4B, 0x03, 0x04}
)

// Subtitle is a single subtitle entry in legendas.tv
type Subtitle struct {
	subtitle subtitleEntry

	c *http.Client // Used for downloading the subtitle file

	language language.Tag        // The language of this subtitle
	target   *sublime.FileTarget // The file this subtitle targets
}

// SubtitlePack is a subtitle entry that is marked as being a pack,
// therefore containing many subtitles
type SubtitlePack struct {
	subtitle subtitleEntry // The pack entry
}

// One subtitle inside a subtitle pack
type SubtitlePackItem struct {
	name     string // This item's filename
	contents []byte // This item's file contents

	language language.Tag        // The language of this subtitle
	target   *sublime.FileTarget // The file this subtitle targets
}

func (s *SubtitlePack) downloadPack(c *http.Client) (archiver.Reader, error) {
	r, err := s.subtitle.DownloadContents(c)
	if err != nil {
		return nil, err
	}

	pack, err := ioutil.ReadAll(r)
	r.Close()
	if err != nil {
		return nil, err
	}

	var archive archiver.Reader

	if bytes.HasPrefix(pack, rar15MagicNumber) || bytes.HasPrefix(pack, rar50MagicNumber) {
		archive = archiver.NewRar()
	} else if bytes.HasPrefix(pack, zipMagicNumber) {
		archive = archiver.NewZip()
	} else {
		return nil, fmt.Errorf(`legendastv: unknown archive format for subtitle "%s"`, s.subtitle.Title)
	}

	err = archive.Open(bytes.NewReader(pack), int64(len(pack)))
	if err != nil {
		return nil, err
	}

	return archive, nil
}

// GetSubtitles downloads and extracts the archive containing the pack,
// and returns each file as a subtitle
func (s *SubtitlePack) GetSubtitles(c *http.Client) ([]SubtitlePackItem, error) {
	archive, err := s.downloadPack(c)
	if err != nil {
		return nil, err
	}
	defer archive.Close()

	res := []SubtitlePackItem{}

	for f, err := archive.Read(); err != io.EOF; f, err = archive.Read() {
		if err != nil {
			return nil, err
		}

		name := f.Name()

		if strings.HasSuffix(name, ".srt") || strings.HasSuffix(name, ".ass") {
			contents, err := ioutil.ReadAll(f)
			f.Close()
			if err != nil {
				return nil, err
			}

			res = append(res, SubtitlePackItem{
				name:     f.Name(),
				contents: contents,
			})
		} else {
			f.Close()
		}
	}

	return res, nil
}

// GetFileTarget returns this subtitle's filetarget
func (s SubtitlePackItem) GetFileTarget() *sublime.FileTarget {
	return s.target
}

// GetLang returns this subtitle's language
func (s SubtitlePackItem) GetLang() language.Tag {
	return s.language
}

// GetInfo parses this subtitle's name for information
// TODO: Inherit some of the pack's info, such as release type
func (s SubtitlePackItem) GetInfo() guessit.Information {
	return guessit.Parse(s.name)
}

// Open returns this subtitle's contents as a stream
func (s SubtitlePackItem) Open() (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewReader(s.contents)), nil
}

// GetFileTarget returns this subtitle's filetarget
func (s Subtitle) GetFileTarget() *sublime.FileTarget {
	return s.target
}

// GetLang returns this subtitle's language
func (s Subtitle) GetLang() language.Tag {
	return s.language
}

// GetInfo parses this subtitle's name for information
// TODO: Inherit some of the pack's info, such as release type
func (s Subtitle) GetInfo() guessit.Information {
	return guessit.Parse(s.subtitle.Title)
}

// Open returns this subtitle's contents as a stream
func (s Subtitle) Open() (io.ReadCloser, error) {
	r, err := s.subtitle.DownloadContents(s.c)
	if err != nil {
		return nil, err
	}

	result, err := ioutil.ReadAll(r)
	r.Close()
	if err != nil {
		return nil, err
	}

	var archive archiver.Reader
	if bytes.HasPrefix(result, rar15MagicNumber) || bytes.HasPrefix(result, rar50MagicNumber) {
		archive = archiver.NewRar()
	} else if bytes.HasPrefix(result, zipMagicNumber) {
		archive = archiver.NewZip()
	}

	if archive != nil {
		defer archive.Close()
		err = archive.Open(bytes.NewReader(result), int64(len(result)))
		if err != nil {
			return nil, err
		}

		// Loop for all of the ".srt" and ".ass" files and select the biggest
		biggest := 0
		for f, err := archive.Read(); err != io.EOF; f, err = archive.Read() {
			if err != nil {
				return nil, err
			}

			name := f.FileInfo.Name()
			if strings.HasSuffix(name, ".srt") || strings.HasSuffix(name, ".ass") {
				buff, err := ioutil.ReadAll(f.ReadCloser)

				if err != nil {
					f.Close()
					return nil, err
				}

				if len(buff) > biggest {
					biggest = len(buff)
					result = buff
				}
			}

			f.Close()
		}
	}

	return ioutil.NopCloser(bytes.NewReader(result)), nil
}
