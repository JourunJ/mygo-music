package lyrics

import (
	"regexp"
	"strconv"
	"strings"

	"mygo-music/internal/music"
)

var (
	qrcLine = regexp.MustCompile(`^\[(\d+),(\d+)\]`)
	qrcWord = regexp.MustCompile(`\((\d+),(\d+),\d+\)`)
	wordTime = regexp.MustCompile(`<((?:\d+):\d{2}(?:[.:]\d+)?)>`)
	lineTime = regexp.MustCompile(`\[(\d+):(\d{2})(?:[.:](\d+))?\]`)
)

// ParseTimed reads word-timed local lyrics. It accepts QRC's millisecond
// offsets and enhanced LRC/Folia files whose words are introduced by <time>.
// It returns false when the text has no recognizable word timings.
func ParseTimed(text string) (*music.Lyrics, bool) {
	var lines []music.LyricLine
	for _, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if m := qrcLine.FindStringSubmatchIndex(raw); len(m) > 0 {
			start, _ := strconv.ParseInt(raw[m[2]:m[3]], 10, 64)
			duration, _ := strconv.ParseInt(raw[m[4]:m[5]], 10, 64)
			matches := qrcWord.FindAllStringSubmatchIndex(raw[m[1]:], -1)
			if len(matches) == 0 { continue }
			body := raw[m[1]:]
			line := music.LyricLine{Start: start, End: start + duration}
			for i, w := range matches {
				wordStart, _ := strconv.ParseInt(body[w[2]:w[3]], 10, 64)
				wordDuration, _ := strconv.ParseInt(body[w[4]:w[5]], 10, 64)
				from := w[1]
				to := len(body)
				if i+1 < len(matches) { to = matches[i+1][0] }
				wordText := body[from:to]
				if wordText != "" {
					line.Words = append(line.Words, music.LyricWord{Start: wordStart, End: wordStart + wordDuration, Text: wordText})
					line.Text += wordText
				}
			}
			if len(line.Words) > 0 { lines = append(lines, line) }
			continue
		}
		matches := wordTime.FindAllStringSubmatchIndex(raw, -1)
		if len(matches) == 0 { continue }
		line := music.LyricLine{}
		// A leading LRC timestamp is the line start; otherwise the first
		// word timestamp supplies it (as in Folia exports).
		if m := lineTime.FindStringSubmatch(raw); len(m) > 0 {
			line.Start = parseClock(m[1], m[2], m[3])
		}
		for i, w := range matches {
			start := parseWordClock(raw[w[2]:w[3]])
			from, to := w[1], len(raw)
			if i+1 < len(matches) { to = matches[i+1][0] }
			wordText := raw[from:to]
			if wordText == "" { continue }
			end := start + 350
			if i+1 < len(matches) {
				end = parseWordClock(raw[matches[i+1][2]:matches[i+1][3]])
			}
			line.Words = append(line.Words, music.LyricWord{Start: start, End: end, Text: wordText})
			line.Text += wordText
		}
		if len(line.Words) > 0 {
			if line.Start == 0 { line.Start = line.Words[0].Start }
			line.End = line.Words[len(line.Words)-1].End
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 { return nil, false }
	return &music.Lyrics{Lines: lines}, true
}

func parseWordClock(s string) int64 {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 { return 0 }
	return parseClock(parts[0], strings.SplitN(parts[1], ".", 2)[0], fractionPart(parts[1]))
}

func fractionPart(s string) string {
	if i := strings.IndexAny(s, ".:"); i >= 0 { return s[i+1:] }
	return ""
}

func parseClock(minute, second, fraction string) int64 {
	m, _ := strconv.ParseInt(minute, 10, 64)
	s, _ := strconv.ParseInt(second, 10, 64)
	var ms int64
	if fraction != "" {
		n, _ := strconv.ParseInt(fraction, 10, 64)
		switch len(fraction) { case 1: ms=n*100; case 2: ms=n*10; default: ms=n/ int64Pow10(len(fraction)-3) }
	}
	return (m*60+s)*1000+ms
}

func int64Pow10(n int) int64 {
	v:=int64(1)
	for range n { v*=10 }
	return v
}

