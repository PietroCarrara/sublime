package legendastv

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

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

	session http.Client
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

	cookies, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	l.session = http.Client{
		Jar: cookies,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

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

	// Page rediects only on success
	if r.StatusCode != http.StatusFound {
		return errors.New("username or password invalid")
	}

	return nil
}

func (l *LegendastvService) GetCandidatesForFiles(files []*sublime.FileTarget, languages []language.Tag) <-chan sublime.SubtitleCandidate {
	// TODO: Implement
	return nil
}
