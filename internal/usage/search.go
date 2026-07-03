package usage

import (
	"context"
	"log"

	nhentai "github.com/Miuzarte/NHentai-go"
	"github.com/Miuzarte/NHentai-go/api"
)

func UsageSearch() {
	const keyword = "耳で恋した同僚〜オナサポ音声オタク女が同僚の声に反応してイキまくり〜"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	page := 0
	sort := nhentai.Sort("") // [nhentai.SORT_POPULAR] | [nhentai.SORT_DATE]
	search, err := nhentai.Search(ctx, keyword, page, sort)
	if err != nil {
		log.Fatalln(err)
	}

	results := api.GalleryListItems(search.Result)

	for _, gallery := range results {
		if gallery.JapaneseTitle != nil {
			log.Println(*gallery.JapaneseTitle)
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
