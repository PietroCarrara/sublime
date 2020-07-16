package legendastv

import (
	"errors"
	"fmt"

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

	// TODO: Login into service

	return nil
}

func (l *LegendastvService) GetCandidatesForFiles(files []*sublime.FileTarget, languages []language.Tag) <-chan sublime.SubtitleCandidate {
	// TODO: Implement
	return nil
}
