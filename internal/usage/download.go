package usage

import (
	"context"
	"log"

	nhentai "github.com/Miuzarte/NHentai-go"
)

func UsageDownload() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const gId = 540994
	gallery, err := nhentai.GetGallery(ctx, gId)
	if err != nil {
		log.Fatalln(err)
	}

	for image, err := range gallery.DownloadPagesIter(ctx) {
		if err != nil {
			log.Println(err)
			break
		}
		log.Println(image.String())
	}

	// download thumbs
	for image, err := range gallery.DownloadThumbsIter(ctx) {
		_, _ = image, err
	}
}
