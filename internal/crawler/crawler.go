package crawler

import (
	"log"
	"time"

	"github.com/GabrielMoody/Cariin/internal/db"
	"github.com/GabrielMoody/Cariin/internal/documents"
)

var baseurl = "https://www.wikipedia.org/"

func Crawl() {
	t := time.Now().Add(30 * time.Second)
	db := db.Init()

	for {
		if time.Now().After(t) {
			log.Println("Finish crawled website!")
			break
		}

		b, err := Fetch(baseurl)
		if err != nil {
			log.Printf("Fetch failed: %v", err)
			continue
		}

		doc, _ := ParseDocument(baseurl, b)
		err = documents.SaveDocument(db, doc)

		if err != nil {
			log.Printf("Db failed: %v", err)
			continue
		}

		links, _ := ExtractLinks(baseurl, b)

		for _, v := range links {
			doc, _ := ParseDocument(v, b)

			err = documents.SaveDocument(db, doc)
			if err != nil {
				log.Printf("Db failed: %v", err)
				continue
			}
		}

	}
}
