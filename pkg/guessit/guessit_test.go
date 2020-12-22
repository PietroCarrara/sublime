package guessit

import (
	"testing"
)

var testCases = map[string]Information{
	"The Walking Dead S05E03 720p HDTV x264-ASAP[ettv]": {
		Title:      "The Walking Dead",
		Season:     5,
		Episode:    3,
		Resolution: "720p",
		Release:    "HDTV",
		VideoCodec: "x264",
		Group:      "ASAP",
	},
	"Battlestar.Galactica.S04E01.BDRip.x264-FGT.mp4": {
		Title:      "Battlestar Galactica",
		Season:     4,
		Episode:    1,
		Release:    "BDRip",
		VideoCodec: "x264",
		Group:      "FGT",
		Container:  "mp4",
	},
	"Alien.1979.REMASTERED.THEATRICAL.1080p.BluRay.H264.AAC-RARBG": {
		Title:      "Alien",
		Year:       1979,
		Resolution: "1080p",
		Release:    "BluRay",
		VideoCodec: "H264",
		AudioCodec: "AAC",
		Group:      "RARBG",
	},
	"Die.Hard.1988.1080p.BluRay.H264.AAC-RARBG": {
		Title:      "Die Hard",
		Year:       1988,
		Resolution: "1080p",
		Release:    "BluRay",
		VideoCodec: "H264",
		AudioCodec: "AAC",
		Group:      "RARBG",
	},
	"Hercules (2014) 1080p BrRip H264 - YIFY.avi": {
		Title:      "Hercules",
		Year:       2014,
		Resolution: "1080p",
		Release:    "BrRip",
		VideoCodec: "H264",
		Group:      "YIFY",
		Container:  "avi",
	},
	"The Big Bang Theory S08E06 HDTV XviD-LOL [eztv]": {
		Title:      "The Big Bang Theory",
		Season:     8,
		Episode:    6,
		Release:    "HDTV",
		VideoCodec: "XviD",
		Group:      "LOL",
	},
	"Brave.2012.R5.DVDRip.XViD.LiNE-UNiQUE": {
		Title:      "Brave",
		Year:       2012,
		Region:     "R5",
		Release:    "DVDRip",
		VideoCodec: "XViD",
		AudioCodec: "LiNE",
		Group:      "UNiQUE",
	},
	"Blade.Runner.1982.DC.Remastered.XviD.AC3-WAF": {
		Title:        "Blade Runner",
		Year:         1982,
		DirectorsCut: true,
		VideoCodec:   "XviD",
		AudioCodec:   "AC3",
		Group:        "WAF",
	},
	"Blade.Runner.2049.2017.4K.UltraHD.BluRay.2160p.x264.TrueHD.Atmos.7.1.AC3-POOP": {
		Title:      "Blade Runner 2049",
		Year:       2017,
		Release:    "BluRay",
		Resolution: "2160p",
		VideoCodec: "x264",
		AudioCodec: "AC3",
		Group:      "POOP",
	},
	"Terminator.2.Judgment.Day.1991.Extended.REMASTERED.1080p.BluRay.H264.AAC.READ.NFO-RARBG": {
		Title:      "Terminator 2 Judgment Day",
		Year:       1991,
		Extended:   true,
		Resolution: "1080p",
		Release:    "BluRay",
		VideoCodec: "H264",
		AudioCodec: "AAC",
		Group:      "RARBG",
	},
	"Greyhound.2020.1080p.WEBRip.x264-RARBG": {
		Title:      "Greyhound",
		Year:       2020,
		Resolution: "1080p",
		Release:    "WEBRip",
		VideoCodec: "x264",
		Group:      "RARBG",
	},
	"[720pMkv.Com]_sons.of.anarchy.s05e10.480p.BluRay.x264-GAnGSteR": {
		Website:    "720pMkv.Com",
		Title:      "sons of anarchy",
		Season:     5,
		Episode:    10,
		Resolution: "480p",
		Release:    "BluRay",
		VideoCodec: "x264",
		Group:      "GAnGSteR",
	},
	"Marvel's.Agents.of.S.H.I.E.L.D.S02E01.Shadows.1080p.WEB-DL.DD5.1": {
		Title:      "Marvel's Agents of S.H.I.E.L.D",
		Season:     2,
		Episode:    1,
		Resolution: "1080p",
		Release:    "WEB-DL",
		AudioCodec: "DD5.1",
	},
	"Interstellar.2014.1080p.BluRay.H264.AAC-RARBG": {
		Title:      "Interstellar",
		Year:       2014,
		Resolution: "1080p",
		Release:    "BluRay",
		VideoCodec: "H264",
		AudioCodec: "AAC",
		Group:      "RARBG",
	},
	"The Missing 1x01 Pilot HDTV x264-FoV [eztv]": {
		Title:      "The Missing",
		Season:     1,
		Episode:    1,
		Release:    "HDTV",
		VideoCodec: "x264",
		Group:      "FoV",
	},
	"Battlestar.Galactica.S04E19E20E21.EXTENDED.BDRip.x264-FGT.pt-BR": {
		Title:      "Battlestar Galactica",
		Season:     4,
		Episode:    19,
		Extended:   true,
		Release:    "BDRip",
		VideoCodec: "x264",
		Group:      "FGT",
	},
	"THX.1138.1971.Directors.Cut.1080p.BluRay.H264.AAC-RARBG": {
		Title:        "THX 1138",
		Year:         1971,
		DirectorsCut: true,
		Resolution:   "1080p",
		Release:      "BluRay",
		VideoCodec:   "H264",
		AudioCodec:   "AAC",
		Group:        "RARBG",
	},
	"E.T.The.Extra-Terrestrial.1982.1080p.BluRay.H264.AAC-RARBG.mp4": {
		Title:      "E.T. The Extra-Terrestrial",
		Year:       1982,
		Resolution: "1080p",
		Release:    "BluRay",
		VideoCodec: "H264",
		AudioCodec: "AAC",
		Group:      "RARBG",
		Container:  "mp4",
	},
	"Apocalypse.Now.1979.Theatrical.REMASTERED.1080p.BluRay.H264.AAC-RARBG.mp4": {
		Title:      "Apocalypse Now",
		Year:       1979,
		Theatrical: true,
		Remastered: true,
		Resolution: "1080p",
		Release:    "BluRay",
		VideoCodec: "H264",
		AudioCodec: "AAC",
		Group:      "RARBG",
		Container:  "mp4",
	},
}

func TestParse(t *testing.T) {
	t.Parallel()

	for filename, target := range testCases {
		value := Parse(filename)

		if !assertEqualsInformation(t, filename, target, value) {
			t.Logf(`(case: "%s") value.Rest => %#v`, filename, value.Rest)
		}
	}
}

func assertEqualsInformation(t *testing.T, testcase string, target, value Information) bool {
	hasError := false

	if target.Title != value.Title {
		t.Errorf(`(case: "%s") Expected "Title" to be %#v, but got %#v`, testcase, target.Title, value.Title)
		hasError = true
	}
	if target.Season != value.Season {
		t.Errorf(`(case: "%s") Expected "Season" to be %#v, but got %#v`, testcase, target.Season, value.Season)
		hasError = true
	}
	if target.Episode != value.Episode {
		t.Errorf(`(case: "%s") Expected "Episode" to be %#v, but got %#v`, testcase, target.Episode, value.Episode)
		hasError = true
	}
	if target.Year != value.Year {
		t.Errorf(`(case: "%s") Expected "Year" to be %#v, but got %#v`, testcase, target.Year, value.Year)
		hasError = true
	}
	if target.Resolution != value.Resolution {
		t.Errorf(`(case: "%s") Expected "Resolution" to be %#v, but got %#v`, testcase, target.Resolution, value.Resolution)
		hasError = true
	}
	if target.Release != value.Release {
		t.Errorf(`(case: "%s") Expected "Release" to be %#v, but got %#v`, testcase, target.Release, value.Release)
		hasError = true
	}
	if target.VideoCodec != value.VideoCodec {
		t.Errorf(`(case: "%s") Expected "VideoCodec" to be %#v, but got %#v`, testcase, target.VideoCodec, value.VideoCodec)
		hasError = true
	}
	if target.AudioCodec != value.AudioCodec {
		t.Errorf(`(case: "%s") Expected "AudioCodec" to be %#v, but got %#v`, testcase, target.AudioCodec, value.AudioCodec)
		hasError = true
	}
	if target.Group != value.Group {
		t.Errorf(`(case: "%s") Expected "Group" to be %#v, but got %#v`, testcase, target.Group, value.Group)
		hasError = true
	}
	if target.Region != value.Region {
		t.Errorf(`(case: "%s") Expected "Region" to be %#v, but got %#v`, testcase, target.Region, value.Region)
		hasError = true
	}
	if target.Extended != value.Extended {
		t.Errorf(`(case: "%s") Expected "Extended" to be %#v, but got %#v`, testcase, target.Extended, value.Extended)
		hasError = true
	}
	if target.DirectorsCut != value.DirectorsCut {
		t.Errorf(`(case: "%s") Expected "DirectorsCut" to be %#v, but got %#v`, testcase, target.DirectorsCut, value.DirectorsCut)
		hasError = true
	}
	if target.Hardcoded != value.Hardcoded {
		t.Errorf(`(case: "%s") Expected "Hardcoded" to be %#v, but got %#v`, testcase, target.Hardcoded, value.Hardcoded)
		hasError = true
	}
	if target.Proper != value.Proper {
		t.Errorf(`(case: "%s") Expected "Proper" to be %#v, but got %#v`, testcase, target.Proper, value.Proper)
		hasError = true
	}
	if target.Container != value.Container {
		t.Errorf(`(case: "%s") Expected "Container" to be %#v, but got %#v`, testcase, target.Container, value.Container)
		hasError = true
	}
	if target.Repack != value.Repack {
		t.Errorf(`(case: "%s") Expected "Repack" to be %#v, but got %#v`, testcase, target.Repack, value.Repack)
		hasError = true
	}
	if target.Widescreen != value.Widescreen {
		t.Errorf(`(case: "%s") Expected "Widescreen" to be %#v, but got %#v`, testcase, target.Widescreen, value.Widescreen)
		hasError = true
	}
	if target.Website != value.Website {
		t.Errorf(`(case: "%s") Expected "Website" to be %#v, but got %#v`, testcase, target.Website, value.Website)
		hasError = true
	}
	if target.Unrated != value.Unrated {
		t.Errorf(`(case: "%s") Expected "Unrated" to be %#v, but got %#v`, testcase, target.Unrated, value.Unrated)
		hasError = true
	}
	if target.Size != value.Size {
		t.Errorf(`(case: "%s") Expected "Size" to be %#v, but got %#v`, testcase, target.Size, value.Size)
		hasError = true
	}
	if target.ThreeD != value.ThreeD {
		t.Errorf(`(case: "%s") Expected "ThreeD" to be %#v, but got %#v`, testcase, target.ThreeD, value.ThreeD)
		hasError = true
	}

	return !hasError
}
