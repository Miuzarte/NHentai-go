package usage

import (
	"context"
	"log"

	nhentai "github.com/Miuzarte/NHentai-go"
)

func UsageSearch() {
	const keyword = "耳で恋した同僚〜オナサポ音声オタク女が同僚の声に反応してイキまくり〜"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	page := 0
	sort := "" // [nhentai.SORT_POPULAR] | [nhentai.SORT_DATE]
	search, err := nhentai.Search(ctx, keyword, page, sort)
	if err != nil {
		log.Fatalln(err)
	}

	results := search.Result

	for _, gallery := range results {
		log.Println(gallery.Title)
		// all results are a complete gallery,
		// no more requests needed.
		// W for NHentai
		for image, err := range gallery.DownloadThumbsIter(ctx) {
			if err != nil {
				log.Fatalln(err)
			}
			log.Println(image.String())
		}
	}

	// download covers
	for image, err := range results.DownloadCoversIter(ctx) {
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(image.String())
	}
}
