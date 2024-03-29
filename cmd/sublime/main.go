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
	"github.com/agnivade/levenshtein"
	"github.com/pkg/errors"
	"golang.org/x/text/language"

	"github.com/jeandeaual/go-locale"

	// Implemented services:
	_ "github.com/PietroCarrara/sublime/pkg/sublime/services/opensubtitles"
)

type releaseType int

const (
	bluray releaseType = iota
	hdtv
	cam
	dvd
	web

	unknown
)

var argLangList = flag.String("languages", "", "comma-separated language list for subtitles")
var argServiceList = flag.String("services", "", "comma-separated service list for subtitles")
var argConfigList = flag.String("config", "", `space-separated list of config values to set in the form service.option=my\ value`)
var argLangNames = flag.String("lnames", "", "comma-separated list of languages to rename in the output (example: pt-Br=pt)")

func main() {
	log.SetFlags(log.Llongfile)

	flag.Parse()

	languages := getLanguages(*argLangList)
	lnames, err := getLangNames(languages, *argLangNames)
	if err != nil {
		log.Fatal(err)
	}

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

	for i, s := range services {
		err := s.Initialize()
		if err != nil {
			log.Println(fmt.Errorf("%s: %s", s.GetName(), err))
			services[i] = nil
		}
	}

	chans := make([]<-chan sublime.SubtitleCandidate, len(services))
	for i := range chans {
		if services[i] != nil {
			chans[i] = services[i].GetCandidatesForFiles(targets, languages)
		}
	}

	channel := unifyChannels(chans)

	best := make(map[*sublime.FileTarget]map[language.Tag]sublime.SubtitleCandidate)
	count := 0
	for sub := range channel {
		count++

		// If we're in a interactive shell
		if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
			fmt.Printf("\rEvaluating %d subtitles...", count)
		}

		f := sub.GetFileTarget()
		l := sub.GetLang()
		if best[f] == nil {
			best[f] = make(map[language.Tag]sublime.SubtitleCandidate)
		}
		if best[f][l] == nil || greater(f.GetInfo(), sub, best[f][l]) {
			best[f][l] = sub
		}
	}
	// If we're not in a interactive shell
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		fmt.Printf("Evaluating %d subtitles...", count)
	}
	fmt.Println()

	for _, f := range targets {
		for _, l := range languages {
			if sub, ok := best[f][l]; ok {
				stream, err := sub.Open()
				if err != nil {
					log.Printf(`could not download subtitle for "%s": %s`, f, err)
					if stream != nil {
						stream.Close()
					}
					fmt.Printf("%s: ✗\n", f)
					continue
				}
				f.SaveSubtitle(stream, lnames[sub.GetLang()], sub.GetFormatExtension())
				stream.Close()
				fmt.Printf("%s [%s]: ✓\n", f, l)
			} else {
				fmt.Printf("%s [%s]: ✗\n", f, l)
			}
		}
	}
}

func getLanguages(langs string) []language.Tag {
	langList := strings.FieldsFunc(langs, func(r rune) bool {
		return r == ','
	})

	var languages []language.Tag
	if len(langList) == 0 {
		loc, err := locale.GetLocales()
		if err != nil {
			log.Fatal(
				"locale: could not determine the languages to download. " +
					"Please, indicate a language using '-language=en-US'.",
			)
		}
		langList = loc
	}

	languages = make([]language.Tag, len(langList))
	for i, l := range langList {
		languages[i] = language.MustParse(l)
	}

	return languages
}

func getLangNames(langs []language.Tag, lnames string) (map[language.Tag]string, error) {
	res := make(map[language.Tag]string)

	for _, l := range langs {
		res[l] = l.String()
	}

	if len(lnames) == 0 {
		return res, nil
	}

	for _, i := range strings.Split(lnames, ",") {
		parts := strings.Split(i, "=")
		if len(parts) != 2 {
			return nil, errors.Errorf(`invalid expression "%s"`, i)
		}
		lang := language.MustParse(parts[0])
		res[lang] = parts[1]
	}

	return res, nil
}

var videoRegex = regexp.MustCompile(`(?i)(wmv|mov|webm|mkv|avi|mp4)$`)

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
func greater(target guessit.Information, subA, subB sublime.SubtitleCandidate) bool {
	a := subA.GetInfo()
	b := subB.GetInfo()

	distance := func(a, b string) int {
		return levenshtein.ComputeDistance(strings.ToLower(a), strings.ToLower(b))
	}

	// Release type is the greatest factor, if the target is not a specific version
	if !target.Extended && !target.DirectorsCut && !target.Theatrical {
		if p(target.Release) == p(a.Release) && p(target.Release) != p(b.Release) {
			return true
		}
		if p(target.Release) == p(b.Release) && p(target.Release) != p(a.Release) {
			return false
		}
	}

	if target.Extended == a.Extended && target.Extended != b.Extended {
		return true
	}
	if target.Extended == b.Extended && target.Extended != a.Extended {
		return false
	}

	if target.Theatrical == a.Theatrical && target.Theatrical != b.Theatrical {
		return true
	}
	if target.Theatrical == b.Theatrical && target.Theatrical != a.Theatrical {
		return false
	}

	if target.DirectorsCut == a.Theatrical && target.DirectorsCut != b.DirectorsCut {
		return true
	}
	if target.DirectorsCut == b.Theatrical && target.DirectorsCut != a.DirectorsCut {
		return false
	}

	if target.Remastered == a.Remastered && target.Remastered != b.Remastered {
		return true
	}
	if target.Remastered == b.Remastered && target.Remastered != a.Remastered {
		return false
	}

	if subA.GetService() == subB.GetService() {
		if subA.GetRanking() > subB.GetRanking() {
			return true
		} else if subA.GetRanking() < subB.GetRanking() {
			return false
		}
	}

	if distance(target.Title, a.Title) < distance(target.Title, b.Title) {
		return true
	}

	return false
}

// alias to parseRelease
func p(s string) releaseType {
	return parseRelease(s)
}

func parseRelease(t string) releaseType {
	t = strings.ToLower(t)

	switch t {
	case "cam-rip",
		"cam",
		"hdcam":
		return cam

	case "dvdr",
		"dvdrip",
		"dvd-full",
		"full-rip",
		"iso rip",
		"lossless rip",
		"untouched rip",
		"dvd-5",
		"dvd-9":
		return dvd

	case "dsr",
		"dsrip",
		"satrip",
		"dthrip",
		"dvbrip",
		"hdtv",
		"pdtv",
		"dtvrip",
		"tvrip",
		"hdtvrip":
		return hdtv

	case "webdl",
		"web dl",
		"web-dl",
		"hdrip",
		"web-dlrip",
		"webrip",
		"web rip",
		"web-rip",
		"web",
		"web-cap",
		"webcap",
		"web cap",
		"hc",
		"hd-rip":
		return web

	case "blu-ray",
		"bluray",
		"blu ray",
		"bdrip",
		"brip",
		"brrip",
		"bdmv",
		"bdr",
		"bd25",
		"bd50",
		"bd5",
		"bd9":
		return bluray

	default:
		return unknown
	}
}

func unifyChannels(channels []<-chan sublime.SubtitleCandidate) <-chan sublime.SubtitleCandidate {
	res := make(chan sublime.SubtitleCandidate)
	totalOpenChannels := len(channels)

	for _, c := range channels {
		if c == nil {
			totalOpenChannels--
			continue
		}
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
					}
					return
				}
			}
		}()
	}

	return res
}
