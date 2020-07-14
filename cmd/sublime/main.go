package main

// "github.com/PietroCarrara/gue"

/**
interface SubtitleCandidate {
	GetInfo()
	GetFileTarget()

	GetStream() io.Reader
}
**/

func main() {
	// lang := ...
	// target := ...
	// services := serviceListOrAll(services.split(','))

	// switch target.type:
	// case Directory:
	// files := target.GetSubtitleFiles()
	// case File:
	// files := [target]

	// chans := make([]chan SubtitleCandidate, len(services))
	// for i in chans:
	//     chans[i] = services[i].getCandidatesForFiles(files, lang)

	// channel := unifyChannels(chans)

	// best := map[FileTarget]*SubtitleCandidate
	// for sub in channel:
	//     f := sub.GetFileTarget()
	//     if best[f] == nil || greater(f.GetInfo(), sub.GetInfo(), best[f].GetInfo()) {
	//          best[f] = &sub
	//     }

	// for f, sub := range best{
	// stream = sub.GetStream()
	// defer stream.Close()
	// f.SaveSubtitle(stream)
	//}
}

/**
func greater(target, a, b Information) int {
	// Returns wheter a > b when matching against target
}

func unifyChannels([]chan SubtitleCandidate) chan-> SubtitleCandidate {
	for chan...
		go select sub, ok := <-c {
			if ok {
				res <- sub
			} else {
				totalOpenChannels--
				if totalOpenChannels <= 0 {
					close(res)
				}
			}
		}
}
**/
