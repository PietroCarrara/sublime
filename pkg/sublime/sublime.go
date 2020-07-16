package sublime

import (
	"io"
	"path"

	"github.com/PietroCarrara/sublime/pkg/guessit"
	"golang.org/x/text/language"
)

// Services is a map keying a service name to it's implementation
var Services = map[string]Service{}

// SubtitleCandidate represents a subtitle that will be evaluated
// and compared to others, to download the best one possible
type SubtitleCandidate interface {
	GetFileTarget() *FileTarget
	GetLang() language.Tag
	GetInfo() guessit.Information
	GetStream() io.ReadCloser
}

// Service knows how to get candidates for FileTargets and Languages
type Service interface {
	// Returns a string identifying this service. Should be all lowercase
	GetName() string
	// For each FileTarget, returns candidates of all of the possible languages
	GetCandidatesForFiles([]*FileTarget, []language.Tag) <-chan SubtitleCandidate
	// Configure values. No costly/long operations should be performed
	SetConfig(name, value string) error
	// Initialize the service
	Initialize() error
}

// FileTarget represents a video that contains information
// and can be sutitled
type FileTarget struct {
	path string
}

// NewFileTarget creates a new FileTarget
func NewFileTarget(path string) *FileTarget {
	return &FileTarget{
		path: path,
	}
}

// GetInfo tries to extract information from a file
func (f FileTarget) GetInfo() guessit.Information {
	_, name := path.Split(f.path)
	return guessit.Parse(name)
}

// SaveSubtitle saves a subtitle next to the video file
func (f FileTarget) SaveSubtitle(r io.Reader, lang language.Tag) error {
	// TODO: Implement
	return nil
}
