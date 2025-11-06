package NHentai

import "testing"

const (
	testId      = 540994
	testMediaId = "3138775"
	testTagId   = 8050
)

func TestGetGallery(t *testing.T) {
	g, err := GetGallery(t.Context(), testId)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", g)
	// for i, url := range g.GetPageUrlsIter() {
	// 	t.Logf("%d: %s", i, url)
	// }
	// g.Related(t.Context())
}

func TestSearchTagged(t *testing.T) {
	s, err := SearchTagged(t.Context(), testTagId, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", s)
}
