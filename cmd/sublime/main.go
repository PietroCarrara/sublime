package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PietroCarrara/sublime/pkg/sublime"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
)

var argLangList = flag.String("language", "", "comma-separated language list for subtitles")
var argServiceList = flag.String("service", "", "comma-separated service list for subtitles")

func main() {

	flag.Parse()

	languages := getLanguages(*argLangList)
	log.Printf("%v\n", languages)

	targets, err := getTargets(flag.Arg(flag.NArg() - 1))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v\n", targets)

	services, err := getServicesOrAll(*argServiceList)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v\n", services)

	// chans := make([]chan SubtitleCandidate, len(services))
	// for i in chans:
	//     chans[i] = services[i].getCandidatesForFiles(files, lang)

	// channel := unifyChannels(chans)

	// best := map[FileTarget]*SubtitleCandidate
	// for sub in channel:
	//     f := sub.GetFileTarget()
	//     if best[f] == nil || greater(f.GetInfo(), sub.GetInfo(), best[f].GetInfo()) {
	//          best[f] = &sub
	//     }

	// for f, sub := range best{
	// stream = sub.GetStream()
	// defer stream.Close()
	// f.SaveSubtitle(stream)
	//}
}

func getLanguages(langs string) []language.Tag {
	langList := strings.FieldsFunc(langs, func(r rune) bool {
		return r == ','
	})

	var languages []language.Tag
	if len(langList) > 0 {
		languages = make([]language.Tag, len(langList))
		for i, l := range langList {
			languages[i] = language.MustParse(l)
		}
	} else {
		// TODO: Use system locale
		panic("not yet implemented!")
	}

	return languages
}

var videoRegex = regexp.MustCompile(`(?i)(mkv|avi|mp4)$`)

func getTargets(path string) ([]*sublime.FileTarget, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// List files and select all videos
	if stat.IsDir() {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}

		res := make([]*sublime.FileTarget, 0, len(files))
		for _, f := range files {
			name := f.Name()

			if videoRegex.MatchString(name) {
				res = append(res, sublime.NewFileTarget(filepath.Join(path, name)))
			}
		}
		return res, nil
	}

	return []*sublime.FileTarget{sublime.NewFileTarget(path)}, nil
}

func getServicesOrAll(services string) ([]sublime.Service, error) {
	serviceList := strings.FieldsFunc(services, func(r rune) bool {
		return r == ','
	})

	if len(serviceList) > 0 {
		res := make([]sublime.Service, len(serviceList))
		var ok bool
		for i, s := range serviceList {
			res[i], ok = sublime.Services[s]
			if !ok {
				return nil, errors.Errorf(`could not locate service "%s"`, s)
			}
		}
		return res, nil
	}

	res := make([]sublime.Service, len(sublime.Services))
	i := 0
	for _, s := range sublime.Services {
		res[i] = s
		i++
	}
	return res, nil
}

/**
func greater(target, a, b Information) int {
	// Returns wheter a > b when matching against target
}

func unifyChannels([]chan SubtitleCandidate) chan-> SubtitleCandidate {
	for chan...
		go select sub, ok := <-c {
			if ok {
				res <- sub
			} else {
				totalOpenChannels--
				if totalOpenChannels <= 0 {
					close(res)
				}
			}
		}
}
**/
