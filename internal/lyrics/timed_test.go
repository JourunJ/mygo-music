package lyrics

import "testing"

// QRC word offsets and enhanced LRC word timestamps must survive parsing exactly.
func TestParseTimedLocalLyrics(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		starts      []int64
		texts       []string
	}{
		{"qrc", "[24880,2040](24880,280,0)朝(25160,280,0)九(25440,170,0)晚", []int64{24880, 25160, 25440}, []string{"朝", "九", "晚"}},
		{"enhanced", "[00:24.880]<00:24.880>朝<00:25.160>九<00:25.440>晚", []int64{24880, 25160, 25440}, []string{"朝", "九", "晚"}},
		{"folia", "<00:24.880>朝<00:25.160>九<00:25.440>晚", []int64{24880, 25160, 25440}, []string{"朝", "九", "晚"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseTimed(tc.input)
			if !ok || len(got.Lines) != 1 {
				t.Fatalf("ParseTimed returned %v, %+v", ok, got)
			}
			if len(got.Lines[0].Words) != len(tc.starts) {
				t.Fatalf("words: %+v", got.Lines[0].Words)
			}
			for i, w := range got.Lines[0].Words {
				if w.Start != tc.starts[i] || w.Text != tc.texts[i] {
					t.Errorf("word %d = %+v", i, w)
				}
			}
		})
	}
}

func TestParseTimedLineStartAndEnd(t *testing.T) {
	input := `[00:36.817]你不记得上次是什么时候见了[00:43.763]
[00:43.763]在路边？[00:45.143]
[00:45.143]还是在路边搭上的出租车上？[00:52.911]
[00:52.911]你们见了[00:55.866]
[00:55.866]一定有个司机知道吧[01:27.174]
[01:27.174]你不记得上次是什么时候见了[01:34.209]`

	got, ok := ParseTimed(input)
	if !ok {
		t.Fatal("expected line-timed lyrics to parse")
	}
	if len(got.Lines) != 6 {
		t.Fatalf("got %d lines, want 6: %+v", len(got.Lines), got.Lines)
	}
	if got.Lines[0].Start != 36817 || got.Lines[0].End != 43763 {
		t.Errorf("first line timing = %d..%d, want 36817..43763", got.Lines[0].Start, got.Lines[0].End)
	}
	if got.Lines[5].Start != 87174 || got.Lines[5].End != 94209 {
		t.Errorf("last line timing = %d..%d, want 87174..94209", got.Lines[5].Start, got.Lines[5].End)
	}
	if got.Lines[0].Text != got.Lines[5].Text {
		t.Errorf("repeated lyric text was changed: %q != %q", got.Lines[0].Text, got.Lines[5].Text)
	}
}

func TestParseTimedSameTimestampTranslation(t *testing.T) {
	input := `[00:10.000]Original lyric[00:11.000]
[00:10.000]Translated lyric[00:11.000]
[00:12.000]Next line[00:13.000]`

	got, ok := ParseTimed(input)
	if !ok {
		t.Fatal("expected timed lyrics to parse")
	}
	if len(got.Lines) != 2 {
		t.Fatalf("got %d lines, want 2: %+v", len(got.Lines), got.Lines)
	}
	if got.Lines[0].Text != "Original lyric" || got.Lines[0].Translation != "Translated lyric" {
		t.Errorf("translation not retained: %+v", got.Lines[0])
	}
}


func TestParseTimedLeavesPlainLRCToLegacyParser(t *testing.T) {
	input := `[00:01.000]line one
[00:02.000]line two`
	if got, ok := ParseTimed(input); ok {
		t.Fatalf("ordinary LRC should not be intercepted: %+v", got)
	}
}
