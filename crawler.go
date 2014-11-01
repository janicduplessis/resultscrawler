package main

import (
	"fmt"
	"log"

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
			lib.Class{
				Name:  "inf1130",
				Group: "20",
				Year:  "20143",
			},
			lib.Class{
				Name:  "met1105",
				Group: "11",
				Year:  "20143",
			},
			lib.Class{
				Name:  "eco1081",
				Group: "51",
				Year:  "20143",
			},
		},
	}

	crawler := new(crawler.Crawler)
	classes, _ := crawler.Run(usertest)

	for _, class := range classes {
		log.Println("----------------------------------")
		log.Println(fmt.Sprintf("Results %v", class.Name))
		log.Println("----------------------------------")
		for _, res := range class.Results {
			log.Println(res.Name)
			log.Println(fmt.Sprintf("  Result:  %v", res.Result))
			log.Println(fmt.Sprintf("  Average: %v", res.Average))
		}
	}
}
