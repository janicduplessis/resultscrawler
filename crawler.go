package main

import (
	"github.com/janicduplessis/resultscrawler/crawler"
	"github.com/janicduplessis/resultscrawler/lib"
)

func main() {
	usertest := &lib.User{
		Code: "*",
		Nip:  "*",
		Classes: []lib.Class{
			lib.Class{
				Name:  "mat1600",
				Group: "20",
				Year:  "20143",
			},
		},
	}

	crawler := new(crawler.Crawler)
	crawler.Run(usertest)
}
