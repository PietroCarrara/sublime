package legendastv

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"

	"github.com/PietroCarrara/sublime/pkg/sublime"
	"golang.org/x/text/language"
)

func init() {
	l := &LegendastvService{}

	sublime.Services[l.GetName()] = l
}

type LegendastvService struct {
	username string
	password string

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

	for _, title := range titles {
		wait.Add(1)
		go downloadTitle(title, out, l.session, &wait)
	}

	go func() {
		wait.Wait()
		close(out)
	}()

	return out
}

// downloadTitle downloads subtitles for all of the targets, as long as they
// are from the same movie/tv show
func downloadTitle(files []*sublime.FileTarget, out chan<- sublime.SubtitleCandidate, c *http.Client, wait *sync.WaitGroup) {
	seasons := make(map[int][]*sublime.FileTarget)
	ourWait := sync.WaitGroup{}

	for _, file := range files {
		info := file.GetInfo()
		if info.Season == 0 && info.Episode == 0 {
			ourWait.Add(1)
			go downloadMovie(file, out, c, &ourWait)
		} else if info.Season == 0 {
			seasons[1] = append(seasons[1], file)
		} else {
			seasons[info.Season] = append(seasons[info.Season], file)
		}
	}

	for _, files := range seasons {
		ourWait.Add(1)
		go downloadSeason(files, out, c, &ourWait)
	}

	ourWait.Wait()
	wait.Done()
}

// downloadMovie a subtitle for a movie
func downloadMovie(file *sublime.FileTarget, out chan<- sublime.SubtitleCandidate, c *http.Client, wait *sync.WaitGroup) {
	// TODO: Implement
	entry, err := getEntry(c, file.GetInfo())
	if err != nil {
		log.Printf("legendastv: %s", err)
	} else {
		log.Println(*entry)
	}
	wait.Done()
}

// downloadSeason downloads subtitles for many files, as long as they
// are from the same tv show and season
func downloadSeason(files []*sublime.FileTarget, out chan<- sublime.SubtitleCandidate, c *http.Client, wait *sync.WaitGroup) {
	// TODO: Implement
	if len(files) > 0 {
		entry, err := getEntry(c, files[0].GetInfo())
		if err != nil {
			log.Printf("legendastv: %s", err)
		} else {
			log.Println(*entry)
		}
	}
	wait.Done()
}
