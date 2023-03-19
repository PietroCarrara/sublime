# sublime
Golang subtitles downloader

## Usage
`$ sublime -config 'opensubtitles.username=user opensubtitles.password=pass legendastv.username=user legendastv.password=pass legendastv.retriesAllowed=10' -languages pt-Br,en 'Squid.Game.S01.KOREAN.WEBRip.x264-ION10/'`

This command will download subtitles for all the files inside the *Squid Game* directory, in brazilian portuguese and english.
Since no services were specified (with `-services opensubtitles,service2,...`), it'll use all the available ones (only OpenSubtitles for now).

## Extending

The codebase is small and simple, so extending this software should be easy enough. Just add a new service at `pkg/sublime/services/` that implements the
service interface and register it. Reading [opensubtitles.go](pkg/sublime/services/opensubtitles/opensubtitles.go) should give you good guidance.

```golang
func init() {
    s := &MyCoolService{}

    sublime.Services[s.GetName()] = s
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
```
