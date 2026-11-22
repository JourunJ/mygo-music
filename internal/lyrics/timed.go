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
	lineTime = regexp.MustCompile(`^\[(\d+):(\d{2})(?:[.:](\d+))?\]`)
	endTime = regexp.MustCompile(`\[(\d+):(\d{2})(?:[.:](\d+))?\]\s*$`)
)

// ParseTimed reads local lyrics with line or word timings. It accepts QRC,
// enhanced LRC/Folia, and LRC lines with an explicit ending timestamp.
func ParseTimed(text string) (*music.Lyrics, bool) {
	var lines []music.LyricLine
	byStart := make(map[int64]int)

	appendLine := func(line music.LyricLine) {
		if line.Text == "" && len(line.Words) == 0 {
			return
		}
		// Some lyric files put the translation on a second line with the
		// same timestamp. Keep it on the original line instead of dropping it.
		if i, ok := byStart[line.Start]; ok && line.Translation == "" {
			if lines[i].Translation == "" && line.Text != "" {
				lines[i].Translation = line.Text
			}
			return
		}
		byStart[line.Start] = len(lines)
		lines = append(lines, line)
	}

	for _, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if m := qrcLine.FindStringSubmatchIndex(raw); len(m) > 0 {
			start, _ := strconv.ParseInt(raw[m[2]:m[3]], 10, 64)
			duration, _ := strconv.ParseInt(raw[m[4]:m[5]], 10, 64)
			body := raw[m[1]:]
			matches := qrcWord.FindAllStringSubmatchIndex(body, -1)
			if len(matches) == 0 {
				continue
			}
			line := music.LyricLine{Start: start, End: start + duration}
			for i, w := range matches {
				wordStart, _ := strconv.ParseInt(body[w[2]:w[3]], 10, 64)
				wordDuration, _ := strconv.ParseInt(body[w[4]:w[5]], 10, 64)
				from, to := w[1], len(body)
				if i+1 < len(matches) {
					to = matches[i+1][0]
				}
				wordText := body[from:to]
				if wordText != "" {
					line.Words = append(line.Words, music.LyricWord{
						Start: wordStart, End: wordStart + wordDuration, Text: wordText,
					})
					line.Text += wordText
				}
			}
			appendLine(line)
			continue
		}

		startMatch := lineTime.FindStringSubmatchIndex(raw)
		if startMatch == nil {
			// Folia-style lines may start directly with word timestamps.
			if !wordTime.MatchString(raw) {
				continue
			}
		}

		line := music.LyricLine{}
		body := raw
		if startMatch != nil {
			line.Start = parseClock(raw[startMatch[2]:startMatch[3]], raw[startMatch[4]:startMatch[5]], fractionFromMatch(raw, startMatch))
			body = raw[startMatch[1]:]
		}

		// A timestamp at the very end is the line's explicit end, not lyric
		// text. This is the [start]text[end] form used by some LRC files.
		if m := endTime.FindStringSubmatchIndex(body); m != nil {
			line.End = parseClock(body[m[2]:m[3]], body[m[4]:m[5]], fractionFromMatch(body, m))
			body = strings.TrimSpace(body[:m[0]])
		}

		matches := wordTime.FindAllStringSubmatchIndex(body, -1)
		if len(matches) == 0 {
			line.Text = strings.TrimSpace(body)
			if line.Text != "" && startMatch != nil {
				appendLine(line)
			}
			continue
		}

		for i, w := range matches {
			wordStart := parseWordClock(body[w[2]:w[3]])
			from, to := w[1], len(body)
			if i+1 < len(matches) {
				to = matches[i+1][0]
			}
			wordText := body[from:to]
			if wordText == "" {
				continue
			}
			wordEnd := wordStart + 350
			if i+1 < len(matches) {
				wordEnd = parseWordClock(body[matches[i+1][2]:matches[i+1][3]])
			} else if line.End > wordStart {
				wordEnd = line.End
			}
			line.Words = append(line.Words, music.LyricWord{
				Start: wordStart, End: wordEnd, Text: wordText,
			})
			line.Text += wordText
		}
		if len(line.Words) > 0 {
			if line.Start == 0 {
				line.Start = line.Words[0].Start
			}
			if line.End == 0 {
				line.End = line.Words[len(line.Words)-1].End
			}
			appendLine(line)
		}
	}

	if len(lines) == 0 {
		return nil, false
	}
	return &music.Lyrics{Lines: lines}, true
}

// fractionFromMatch returns the optional fractional part of a [mm:ss.xxx] tag.
func fractionFromMatch(s string, m []int) string {
	if m[1] < m[0] || m[1] > len(s) {
		return ""
	}
	tag := s[m[0]:m[1]]
	if i := strings.IndexAny(tag, ".:"); i >= 0 {
		if close := strings.LastIndex(tag, "]"); close > i {
			return tag[i+1 : close]
		}
	}
	return ""
}

func parseWordClock(s string) int64 {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0
	}
	second := parts[1]
	if i := strings.IndexAny(second, ".:"); i >= 0 {
		second = second[:i]
	}
	return parseClock(parts[0], second, fractionPart(parts[1]))
}

func fractionPart(s string) string {
	if i := strings.IndexAny(s, ".:"); i >= 0 {
		return s[i+1:]
	}
	return ""
}

func parseClock(minute, second, fraction string) int64 {
	m, _ := strconv.ParseInt(minute, 10, 64)
	s, _ := strconv.ParseInt(second, 10, 64)
	var ms int64
	if fraction != "" {
		n, _ := strconv.ParseInt(fraction, 10, 64)
		switch len(fraction) {
		case 1:
			ms = n * 100
		case 2:
			ms = n * 10
		default:
			ms = n / int64Pow10(len(fraction)-3)
		}
	}
	return (m*60+s)*1000 + ms
}

func int64Pow10(n int) int64 {
	v := int64(1)
	for range n {
		v *= 10
	}
	return v
}
