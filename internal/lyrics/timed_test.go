package lyrics

import "testing"

// QRC word offsets and enhanced LRC word timestamps must survive parsing exactly.
func TestParseTimedLocalLyrics(t *testing.T) {
	for _, tc := range []struct{name, input string; starts []int64; texts []string}{
		{"qrc", "[24880,2040](24880,280,0)朝(25160,280,0)九(25440,170,0)晚", []int64{24880,25160,25440}, []string{"朝","九","晚"}},
		{"enhanced", "[00:24.880]<00:24.880>朝<00:25.160>九<00:25.440>晚", []int64{24880,25160,25440}, []string{"朝","九","晚"}},
		{"folia", "<00:24.880>朝<00:25.160>九<00:25.440>晚", []int64{24880,25160,25440}, []string{"朝","九","晚"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseTimed(tc.input)
			if !ok || len(got.Lines) != 1 { t.Fatalf("ParseTimed returned %v, %+v", ok, got) }
			if len(got.Lines[0].Words) != len(tc.starts) { t.Fatalf("words: %+v", got.Lines[0].Words) }
			for i, w := range got.Lines[0].Words {
				if w.Start != tc.starts[i] || w.Text != tc.texts[i] { t.Errorf("word %d = %+v", i, w) }
			}
		})
	}
}
