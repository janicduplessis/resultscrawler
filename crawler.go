package main

import (
	"log"

	"github.com/janicduplessis/resultscrawler/config"
	"github.com/janicduplessis/resultscrawler/crawler"
	"github.com/janicduplessis/resultscrawler/lib"
)

func main() {
	// Default config
	conf := &config.ServerConfig{
		ServerURL:  "localhost",
		ServerPort: "9898",
		DbName:     "resultscrawler",
		DbUser:     "resultscrawler",
		DbPassword: "***",
		DbURL:      "localhost",
		DbPort:     "7777",
	}

	config.ReadFile("crawler.json", conf)
	config.Print(conf)

	// Inject dependencies
	store := lib.NewMongoStore()
	userStore := lib.NewUserStoreHandler(store)
	scheduler := crawler.NewScheduler(userStore)

	log.Println("Server started")
	/*err := userStore.Insert(&lib.User{
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
	})
	if err != nil {
		log.Println(err)
	}*/

	scheduler.Start()
	log.Println("Server stopped")

	/*for _, class := range classes {
		log.Println("----------------------------------")
		log.Println(fmt.Sprintf("Results %v", class.Name))
		log.Println("----------------------------------")
		for _, res := range class.Results {
			log.Println(res.Name)
			log.Println(fmt.Sprintf("  Result:  %v", res.Result))
			log.Println(fmt.Sprintf("  Average: %v", res.Average))
		}
	}*/
}
