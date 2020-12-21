package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PietroCarrara/sublime/pkg/guessit"
	"github.com/PietroCarrara/sublime/pkg/sublime"
	"github.com/pkg/errors"
	"golang.org/x/text/language"

	// Implemented services:
	_ "github.com/PietroCarrara/sublime/pkg/sublime/services/legendastv"
)

var argLangList = flag.String("language", "", "comma-separated language list for subtitles")
var argServiceList = flag.String("service", "", "comma-separated service list for subtitles")
var argConfigList = flag.String("config", "", `space-separated list of config values to set in the form service.option=my\ value`)

func main() {
	flag.Parse()

	languages := getLanguages(*argLangList)

	targets, err := getTargets(flag.Arg(flag.NArg() - 1))
	if err != nil {
		log.Fatal(err)
	}

	services, err := getServicesOrAll(*argServiceList)
	if err != nil {
		log.Fatal(err)
	}

	err = configServices(*argConfigList)
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range services {
		err := s.Initialize()
		if err != nil {
			log.Fatal(fmt.Errorf("%s: %s", s.GetName(), err))
		}
	}

	chans := make([]<-chan sublime.SubtitleCandidate, len(services))
	for i := range chans {
		chans[i] = services[i].GetCandidatesForFiles(targets, languages)
	}

	channel := unifyChannels(chans)

	best := make(map[*sublime.FileTarget]sublime.SubtitleCandidate)
	count := 0
	for sub := range channel {
		count++

		// If we're in a interactive shell
		if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
			fmt.Printf("\rEvaluating %d subtitles...", count)
		}

		f := sub.GetFileTarget()
		if best[f] == nil || greater(f.GetInfo(), sub.GetInfo(), best[f].GetInfo()) {
			best[f] = sub
		}
	}
	// If we're not in a interactive shell
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		fmt.Printf("Evaluating %d subtitles...", count)
	}
	fmt.Println()

	for _, f := range targets {
		if sub, ok := best[f]; ok {
			stream, err := sub.Open()
			if err != nil {
				log.Printf(`could not download subtitle for "%s": %s`, f, err)
				if stream != nil {
					stream.Close()
				}
				fmt.Printf("%s: ✗\n", f)
				continue
			}
			defer stream.Close()
			f.SaveSubtitle(stream, sub.GetLang())
			fmt.Printf("%s: ✓\n", f)
		} else {
			fmt.Printf("%s: ✗\n", f)
		}
	}
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

			if f.IsDir() {
				// Recursively scan subdirectories
				sub, err := getTargets(filepath.Join(path, name))
				if err != nil {
					return nil, err
				}
				res = append(res, sub...)
			} else {
				// If it's a file and matches the video regex...
				if videoRegex.MatchString(name) {
					res = append(res, sublime.NewFileTarget(filepath.Join(path, name)))
				}
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

var separationRegex = regexp.MustCompile(`[^\\] `)

func configServices(list string) error {
	list = strings.Trim(list, " \n\r\t")

	if len(list) == 0 {
		return nil
	}

	seps := separationRegex.FindAllStringIndex(list, -1)

	last := 0
	for _, slice := range seps {
		opt := list[last : slice[0]+1]

		if err := setConfig(opt); err != nil {
			return err
		}

		last = slice[1]
	}
	if err := setConfig(list[last:]); err != nil {
		return err
	}

	return nil
}

// Sets a config value in a service using a string in the form:
// service.config=value
func setConfig(config string) error {

	parts := strings.SplitN(config, ".", 2)

	if len(parts) != 2 {
		return fmt.Errorf(`config string "%s" not propperly formatted`, config)
	}

	service := parts[0]
	rest := parts[1]

	parts = strings.SplitN(rest, "=", 2)

	if len(parts) != 2 {
		return fmt.Errorf(`config string "%s" not propperly formatted`, config)
	}

	key := parts[0]
	value := parts[1]

	if s, ok := sublime.Services[service]; ok {
		err := s.SetConfig(key, value)
		if err != nil {
			return fmt.Errorf("%s: %s", service, err)
		}
	} else {
		return fmt.Errorf(`service "%s" was not found`, service)
	}

	return nil
}

// greater returns wether A is a better match than B is
// when compared to target
func greater(target, a, b guessit.Information) bool {
	if target.Extended == a.Extended && target.Extended != b.Extended {
		return true
	}

	if target.Theatrical == a.Theatrical && target.Theatrical != b.Theatrical {
		return true
	}

	if target.DirectorsCut == a.Theatrical && target.DirectorsCut != b.DirectorsCut {
		return true
	}

	if target.Remastered == a.Remastered && target.Remastered != b.Remastered {
		return true
	}

	if l(target.Release) == l(a.Release) && l(target.Release) != l(b.Release) {
		return true
	}

	// TODO: Better classification

	return false
}

// alias to strings.ToLower
func l(s string) string {
	return strings.ToLower(s)
}

func unifyChannels(channels []<-chan sublime.SubtitleCandidate) <-chan sublime.SubtitleCandidate {
	res := make(chan sublime.SubtitleCandidate)
	totalOpenChannels := len(channels)

	for _, c := range channels {
		c := c
		go func() {
			for {
				sub, ok := <-c

				if ok {
					res <- sub
				} else {
					totalOpenChannels--
					if totalOpenChannels <= 0 {
						close(res)
						return
					}
				}
			}
		}()
	}

	return res
}
