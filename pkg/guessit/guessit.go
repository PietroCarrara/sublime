package guessit

import (
	"regexp"
	"strconv"
)

// Patterns from:
// https://github.com/divijbindlish/parse-torrent-name/blob/master/PTN/patterns.py

func init() {
	reSeason = regexp.MustCompile(`(?i)(s?([0-9]{1,2}))[ex]`)
	reEpisode = regexp.MustCompile(`(?i)([ex]([0-9]{2})(?:[^0-9]|$))`)
	reYear = regexp.MustCompile(`(?i)\b([\[\(]?((?:19[0-9]|20[01])[0-9])[\]\)]?)\b`)
	reResolution = regexp.MustCompile(`(?i)\b([0-9]{3,4}p)\b`)
	reQuality = regexp.MustCompile(`(?i)\b((?:PPV\.)?[HP]DTV|(?:HD)?CAM|B[DR]Rip|(?:HD-?)?TS|(?:PPV )?WEB-?DL(?: DVDRip)?|HDRip|DVDRip|DVDRIP|CamRip|W[EB]BRip|BluRay|DvDScr|hdtv|telesync)\b`)
	reVideoCodec = regexp.MustCompile(`(?i)\b(xvid|[hx]\.?26[45])\b`)
	reAudioCodec = regexp.MustCompile(`(?i)\b(MP3|DD5\.?1|Dual[\- ]Audio|LiNE|DTS|AAC[.-]LC|AAC(?:\.?2\.0)?|AC3(?:\.5\.1)?)\b`)
	reGroup = regexp.MustCompile(`(?i)\b(- ?([^-]+(?:-={[^-]+-?$)?))$\b`)
	reRegion = regexp.MustCompile(`(?i)\bR[0-9]\b`)
	reExtended = regexp.MustCompile(`(?i)\b(EXTENDED(:?.CUT)?)\b`)
	reHardcoded = regexp.MustCompile(`(?i)\bHC\b`)
	reProper = regexp.MustCompile(`(?i)\bPROPER\b`)
	reContainer = regexp.MustCompile(`(?i)\bREPACK\b`)
	reRepack = regexp.MustCompile(`(?i)\b(MKV|AVI|MP4)\b`)
	reWidescreen = regexp.MustCompile(`(?i)\bWS\b`)
	reWebsite = regexp.MustCompile(`(?i)^(\[ ?([^\]]+?) ?\])`)
	reUnrated = regexp.MustCompile(`(?i)\bUNRATED\b`)
	reSize = regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?(?:GB|MB))\b`)
	reThreeD = regexp.MustCompile(`(?i)\b3D\b`)
}

var (
	reSeason     *regexp.Regexp
	reEpisode    *regexp.Regexp
	reYear       *regexp.Regexp
	reResolution *regexp.Regexp
	reQuality    *regexp.Regexp
	reVideoCodec *regexp.Regexp
	reAudioCodec *regexp.Regexp
	reGroup      *regexp.Regexp
	reRegion     *regexp.Regexp
	reExtended   *regexp.Regexp
	reHardcoded  *regexp.Regexp
	reProper     *regexp.Regexp
	reContainer  *regexp.Regexp
	reRepack     *regexp.Regexp
	reWidescreen *regexp.Regexp
	reWebsite    *regexp.Regexp
	reUnrated    *regexp.Regexp
	reSize       *regexp.Regexp
	reThreeD     *regexp.Regexp
)

// Information holds data regarding a media release
type Information struct {
	Title      string // Media title. "" if none
	Season     int    // Season number. 0 if none
	Episode    int    // Episode number. 0 if none
	Year       int    // Media release year. 0 if none
	Resolution string // Video mode (1080p, 720i...). "" if none
	Release    string // Release type (BDRip, WEBRip...). "" if none. See https://en.wikipedia.org/wiki/Pirated_movie_release_types
	VideoCodec string // Video codec (XViD, h264...). "" if none
	AudioCodec string // Audio codec (AAC, MP3...). "" if none
	Group      string // Group responsible for the release. "" if none
	Region     string // Media region. "" if none
	Extended   bool   // Is the media a extended version?
	Hardcoded  bool   // Media has hardcoded pixels?
	Proper     bool   // Media was re-released fixing problems in previous releases?
	Container  string // Containter for the media. "" if none
	Repack     bool   // Is the release a repack?
	Widescreen bool   // Is the media widescreen?
	Website    string // Release website. "" if none
	Unrated    bool   // The media hasn't been rated for age restricions
	Size       uint64 // Media size in bytes
	ThreeD     bool   // Is the media 3D?
}

// Parse parses a string extraction information in it
func Parse(str string) (*Information, error) {
	res := Information{}

	// Check if group 2 was found
	seasonMatch := reSeason.FindStringSubmatch(str)
	if len(seasonMatch) >= 3 {
		season, err := strconv.Atoi(seasonMatch[2])
		if err != nil {
			return nil, err
		}
		res.Season = season
	}

	// Check if group 2 was found
	episodeMatch := reEpisode.FindStringSubmatch(str)
	if len(episodeMatch) >= 3 {
		episode, err := strconv.Atoi(episodeMatch[2])
		if err != nil {
			return nil, err
		}
		res.Episode = episode
	}

	yearMatch := reYear.FindString(str)
	if len(yearMatch) > 0 {
		year, err := strconv.Atoi(yearMatch)
		if err != nil {
			return nil, err
		}
		res.Year = year
	}

	resolutionMatch := reResolution.FindString(str)
	if len(resolutionMatch) > 0 {
		res.Resolution = resolutionMatch
	}

	// TODO: parse the rest of the information

	return &res, nil
}
