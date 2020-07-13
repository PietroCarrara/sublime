package guessit

import (
	"regexp"
	"strconv"
)

// Patterns taken (and modified) from:
// https://github.com/divijbindlish/parse-torrent-name/blob/master/PTN/patterns.py

func init() {
	reSeason = regexp.MustCompile(`(?i)(s?([0-9]{1,2}))[ex]`)
	reEpisode = regexp.MustCompile(`(?i)([ex]([0-9]{2})(?:[^0-9]|$))`)
	reYear = regexp.MustCompile(`(?i)\b([\[\(]?(\d{4})[\]\)]?)\b`)
	reResolution = regexp.MustCompile(`(?i)\b([0-9]{3,4}p)\b`)
	reRelease = regexp.MustCompile(`(?i)\b((?:PPV\.)?[HP]DTV|(?:HD)?CAM|B[DR]Rip|(?:HD-?)?TS|(?:PPV )?WEB-?DL(?: DVDRip)?|HDRip|DVDRip|DVDRIP|CamRip|W[EB]BRip|BluRay|DvDScr|hdtv|telesync)\b`)
	reVideoCodec = regexp.MustCompile(`(?i)\b(xvid|[hx]\.?26[45])\b`)
	reAudioCodec = regexp.MustCompile(`(?i)\b(MP3|DD5\.?1|Dual[\- ]Audio|LiNE|DTS|AAC[.-]LC|AAC(?:\.?2\.0)?|AC3(?:\.5\.1)?)\b`)
	reGroup = regexp.MustCompile(`(?i)(- ?([^-\.\[ ]+(?:-={[^-]+-?$)?))`)
	reRegion = regexp.MustCompile(`(?i)\bR[0-9]\b`)
	reExtended = regexp.MustCompile(`(?i)\b(EXTENDED(:?.CUT)?)\b`)
	reDirectorsCut = regexp.MustCompile(`(?i)\b(?:(?:DC)|(?:DIRECTORS.CUT))\b`)
	reHardcoded = regexp.MustCompile(`(?i)\bHC\b`)
	reProper = regexp.MustCompile(`(?i)\bPROPER\b`)
	reContainer = regexp.MustCompile(`(?i)\b(MKV|AVI|MP4)\b`)
	reRepack = regexp.MustCompile(`(?i)\bREPACK\b`)
	reWidescreen = regexp.MustCompile(`(?i)\bWS\b`)
	reWebsite = regexp.MustCompile(`(?i)^(\[ ?([^\]]+?) ?\])`)
	reUnrated = regexp.MustCompile(`(?i)\bUNRATED\b`)
	reSize = regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?(?:GB|MB))\b`)
	reThreeD = regexp.MustCompile(`(?i)\b3D\b`)
}

var (
	reSeason       *regexp.Regexp
	reEpisode      *regexp.Regexp
	reYear         *regexp.Regexp
	reResolution   *regexp.Regexp
	reRelease      *regexp.Regexp
	reVideoCodec   *regexp.Regexp
	reAudioCodec   *regexp.Regexp
	reGroup        *regexp.Regexp
	reRegion       *regexp.Regexp
	reExtended     *regexp.Regexp
	reDirectorsCut *regexp.Regexp
	reHardcoded    *regexp.Regexp
	reProper       *regexp.Regexp
	reContainer    *regexp.Regexp
	reRepack       *regexp.Regexp
	reWidescreen   *regexp.Regexp
	reWebsite      *regexp.Regexp
	reUnrated      *regexp.Regexp
	reSize         *regexp.Regexp
	reThreeD       *regexp.Regexp
)

// Information holds data regarding a media release
type Information struct {
	Title        string // Media title. "" if none
	Season       int    // Season number. 0 if none
	Episode      int    // Episode number. 0 if none
	Year         int    // Media release year. 0 if none
	Resolution   string // Video mode (1080p, 720i...). "" if none
	Release      string // Release type (BDRip, WEBRip...). "" if none. See https://en.wikipedia.org/wiki/Pirated_movie_release_types
	VideoCodec   string // Video codec (XViD, h264...). "" if none
	AudioCodec   string // Audio codec (AAC, MP3...). "" if none
	Group        string // Group responsible for the release. "" if none
	Region       string // Media region. "" if none
	Extended     bool   // Is the media a extended version?
	DirectorsCut bool   // Is the media director's cut version?
	Hardcoded    bool   // Media has hardcoded pixels?
	Proper       bool   // Media was re-released fixing problems in previous releases?
	Container    string // Containter for the media. "" if none
	Repack       bool   // Is the release a repack?
	Widescreen   bool   // Is the media widescreen?
	Website      string // Release website. "" if none
	Unrated      bool   // Has the media not been rated for age restricions?
	Size         string // Media size (900MB, 1.3 GB). "" if none
	ThreeD       bool   // Is the media 3D?
}

// Parse parses a string extraction information in it
func Parse(str string) (*Information, error) {
	res := Information{}

	seasonMatch := reSeason.FindStringSubmatchIndex(str)
	if seasonStr := getNthGroup(str, seasonMatch, 2); seasonStr != "" {
		season, err := strconv.Atoi(seasonStr)
		if err != nil {
			return nil, err
		}
		res.Season = season
	}

	episodeMatch := reEpisode.FindStringSubmatchIndex(str)
	if episodeStr := getNthGroup(str, episodeMatch, 2); episodeStr != "" {
		episode, err := strconv.Atoi(episodeStr)
		if err != nil {
			return nil, err
		}
		res.Episode = episode
	}

	// Find last occurrance of a year
	yearMatchAll := reYear.FindAllStringIndex(str, -1)
	if len(yearMatchAll) > 0 {
		if yearStr := getNthGroup(str, yearMatchAll[len(yearMatchAll)-1], 0); yearStr != "" {
			year, err := strconv.Atoi(yearStr)
			if err != nil {
				return nil, err
			}
			res.Year = year
		}
	}

	resolutionMatch := reResolution.FindStringIndex(str)
	res.Resolution = getNthGroup(str, resolutionMatch, 0)

	releaseMatch := reRelease.FindStringIndex(str)
	res.Release = getNthGroup(str, releaseMatch, 0)

	videoCodecMatch := reVideoCodec.FindStringIndex(str)
	res.VideoCodec = getNthGroup(str, videoCodecMatch, 0)

	audioCodecMatch := reAudioCodec.FindStringIndex(str)
	res.AudioCodec = getNthGroup(str, audioCodecMatch, 0)

	groupMatch := reGroup.FindStringSubmatchIndex(str)
	res.Group = getNthGroup(str, groupMatch, 2)

	regionMatch := reRegion.FindStringIndex(str)
	res.Region = getNthGroup(str, regionMatch, 0)

	// For boolean attributes we just need to find a match
	extendedMatch := reExtended.FindStringIndex(str)
	res.Extended = extendedMatch != nil

	directorsCutMatch := reDirectorsCut.FindStringIndex(str)
	res.DirectorsCut = directorsCutMatch != nil

	hardcodedMatch := reHardcoded.FindStringIndex(str)
	res.Hardcoded = hardcodedMatch != nil

	properMatch := reProper.FindStringIndex(str)
	res.Proper = properMatch != nil

	containerMatch := reContainer.FindStringIndex(str)
	res.Container = getNthGroup(str, containerMatch, 0)

	repackMatch := reRepack.FindStringIndex(str)
	res.Proper = repackMatch != nil

	widescreenMatch := reWidescreen.FindStringIndex(str)
	res.Proper = widescreenMatch != nil

	websiteMatch := reWebsite.FindStringIndex(str)
	res.Website = getNthGroup(str, websiteMatch, 2)

	unratedMatch := reUnrated.FindStringIndex(str)
	res.Unrated = unratedMatch != nil

	sizeMatch := reSize.FindStringIndex(str)
	res.Size = getNthGroup(str, sizeMatch, 0)

	threeDMatch := reThreeD.FindStringIndex(str)
	res.ThreeD = threeDMatch != nil

	return &res, nil
}

// getNthGroup returns the string in the nth capture group
// indexPairs is an array returned by a regexp's FindIndex or similar
// The 0th group is the whole match
// If the requested group doesn't exist, getNthGroup returns ""
func getNthGroup(str string, indexPairs []int, group int) string {
	if group >= len(indexPairs)/2 {
		return ""
	}
	return str[indexPairs[group*2]:indexPairs[group*2+1]]
}
