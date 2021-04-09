package legendastv

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/PietroCarrara/sublime/pkg/sublime"
	"golang.org/x/text/language"
)

var languagesID = map[language.Tag]int{
	language.BrazilianPortuguese: 1,
	language.English:             2,
	language.Spanish:             3,
	language.French:              4,
	language.German:              5,
	language.Japanese:            6,
	language.Danish:              7,
	language.Norwegian:           8,
	language.Swedish:             9,
	language.EuropeanPortuguese:  10,
	language.Czech:               12,
	language.Chinese:             13,
	language.Korean:              14,
	language.Bulgarian:           15,
	language.Italian:             16,
	language.Polish:              17,
}

var languagesTag = map[int]language.Tag{
	1:  language.BrazilianPortuguese,
	2:  language.English,
	3:  language.Spanish,
	4:  language.French,
	5:  language.German,
	6:  language.Japanese,
	7:  language.Danish,
	8:  language.Norwegian,
	9:  language.Swedish,
	10: language.EuropeanPortuguese,
	12: language.Czech,
	13: language.Chinese,
	14: language.Korean,
	15: language.Bulgarian,
	16: language.Italian,
	17: language.Polish,
}

func init() {
	l := &LegendastvService{
		retriesAllowed: 1,
	}

	sublime.Services[l.GetName()] = l
}

type LegendastvService struct {
	username string
	password string

	// number of retries allowed to retry a query to find
	// the show/movie entry on the website
	retriesAllowed int

	session *http.Client
}

func (l *LegendastvService) GetName() string {
	return "legendastv"
}

func (l *LegendastvService) SetConfig(name, value string) error {

	switch name {
	case "username":
		l.username = value
	case "password":
		l.password = value
	case "retriesAllowed":
		var err error
		l.retriesAllowed, err = strconv.Atoi(value)
		return err
	default:
		return fmt.Errorf(`option "%s" was not found`, name)
	}

	return nil
}

func (l *LegendastvService) Initialize() error {
	if l.username == "" || l.password == "" {
		return errors.New("username and password are required")
	}

	// Setup session
	cookies, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	l.session = &http.Client{
		Jar: cookies,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Login on session
	req, err := http.NewRequest("POST", "http://legendas.tv/login", strings.NewReader(url.Values{
		"data[User][username]": {l.username},
		"data[User][password]": {l.password},
	}.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	r, err := l.session.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	// Page rediects only on success
	if r.StatusCode != http.StatusFound {
		return errors.New("username or password invalid")
	}

	return nil
}

func (l *LegendastvService) GetCandidatesForFiles(files []*sublime.FileTarget, languages []language.Tag) <-chan sublime.SubtitleCandidate {

	out := make(chan sublime.SubtitleCandidate)
	wait := sync.WaitGroup{}

	titles := make(map[string][]*sublime.FileTarget)
	for _, file := range files {
		info := file.GetInfo()
		titles[info.Title] = append(titles[info.Title], file)
	}

	for _, lang := range languages {
		if id, ok := languagesID[lang]; ok {
			for _, title := range titles {
				wait.Add(1)
				go downloadTitle(title, id, out, l, &wait)
			}
		} else {
			log.Printf(`legendastv: language "%s" is not supported`, lang)
		}
	}

	go func() {
		wait.Wait()
		close(out)
	}()

	return out
}

// downloadTitle downloads subtitles for all of the targets, as long as they
// are from the same movie/tv show
func downloadTitle(files []*sublime.FileTarget, langID int, out chan<- sublime.SubtitleCandidate, l *LegendastvService, wait *sync.WaitGroup) {
	seasons := make(map[int][]*sublime.FileTarget)
	ourWait := sync.WaitGroup{}

	for _, file := range files {
		info := file.GetInfo()
		if info.Season == 0 && info.Episode == 0 {
			ourWait.Add(1)
			go downloadLegendasTV([]*sublime.FileTarget{file}, langID, out, l, &ourWait)
		} else if info.Season == 0 {
			seasons[1] = append(seasons[1], file)
		} else {
			seasons[info.Season] = append(seasons[info.Season], file)
		}
	}

	for _, files := range seasons {
		ourWait.Add(1)
		go downloadLegendasTV(files, langID, out, l, &ourWait)
	}

	ourWait.Wait()
	wait.Done()
}

// downloadSeason downloads subtitles for many files, as long as they all
// are from the same media in legendas.tv (a movie or a tv show season)
func downloadLegendasTV(files []*sublime.FileTarget, langID int, out chan<- sublime.SubtitleCandidate, l *LegendastvService, wait *sync.WaitGroup) {
	defer wait.Done()

	ourWait := sync.WaitGroup{}
	defer ourWait.Wait()

	if len(files) <= 0 {
		return
	}

	var entries []*mediaEntry
	var err error

	for i := 0; i < l.retriesAllowed; i++ {
		entries, err = getEntries(l.session, files[0].GetInfo())

		if err == nil {
			break
		}
	}

	if err != nil {
		log.Printf("legendastv: %s", err)
		return
	}

	for _, entry := range entries {
		subs, err := entry.ListSubtitles(l.session, subTypeAny, langID)
		if err != nil {
			log.Printf("legendastv: Error while searching subtitles: %s", err)
			return
		}

		for _, subEntry := range subs {
			ourWait.Add(1)
			go func(subEntry subtitleEntry) {
				var err error
				switch subEntry.Type {
				case subTypePack:
					err = downloadSubPack(subEntry, l, out, files, langID)
				case subTypeHighlight:
				case subTypeAny:
					err = downloadSubEntry(subEntry, l, out, files, langID)
				default:
					err = fmt.Errorf("legendastv: Subtitle '%s' has an unknown type", subEntry.Title)
				}
				if err != nil {
					log.Printf("legendastv: %s", err)
				}
				ourWait.Done()
			}(subEntry)
		}

		if len(subs) > 0 {
			break
		}
	}
}

// downloadSubPack Downloads a single subtitle pack for many files, as long as they all
// are from the same media
func downloadSubPack(entry subtitleEntry, l *LegendastvService, out chan<- sublime.SubtitleCandidate, files []*sublime.FileTarget, langID int) error {
	pack := SubtitlePack{
		entry: entry,
	}

	subs, err := pack.GetSubtitles(l.session)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		info := sub.GetInfo()

		for _, file := range files {
			if file.GetInfo().Episode == info.Episode {
				sub.language = languagesTag[langID]
				sub.target = file

				out <- sub
				break
			}
		}
	}

	return nil
}

func downloadSubEntry(entry subtitleEntry, l *LegendastvService, out chan<- sublime.SubtitleCandidate, files []*sublime.FileTarget, langID int) error {
	sub := Subtitle{
		c:        l.session,
		subtitle: entry,
		language: languagesTag[langID],
	}
	info := sub.GetInfo()

	for _, file := range files {

		if file.GetInfo().Episode == info.Episode {
			sub.target = file
			out <- &sub
			break
		}
	}

	return nil
}
