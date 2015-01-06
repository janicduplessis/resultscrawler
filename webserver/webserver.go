package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/janicduplessis/resultscrawler/pkg/crypto"
	"github.com/janicduplessis/resultscrawler/pkg/store/mongo"
	"github.com/janicduplessis/resultscrawler/pkg/tools"
	"github.com/janicduplessis/resultscrawler/pkg/webserver"
)

const (
	configFile = "config.json"
)

var (
	serverPort        = flag.String("port", "", "Server port")
	dbURL             = flag.String("db-url", "", "DB url")
	dbUser            = flag.String("db-user", "", "DB user")
	dbPassword        = flag.String("db-password", "", "DB password")
	dbName            = flag.String("db-name", "", "DB name")
	sessionKey        = flag.String("session-key", "", "Session key")
	aesSecret         = flag.String("aes-secret", "", "AES secret key")
	crawlerServiceURL = flag.String("crawler-url", "", "Crawler webservice url")
)

type config struct {
	ServerPort           string
	Database             *tools.MongoConfig
	AESSecretKey         string // 16 bytes
	SessionKey           string
	CrawlerWebserviceURL string
}

func main() {
	log.SetFlags(log.Lshortfile)

	flag.Parse()
	config := readConfig()
	crypto.Init(config.AESSecretKey)

	// Inject dependencies
	mongoHelper := tools.NewMongoHelper(config.Database)

	userStore := mongo.New(mongoHelper)
	crawlerConfigStore := mongo.New(mongoHelper)
	userResultsStore := mongo.New(mongoHelper)

	server := webserver.NewWebserver(&webserver.Config{
		UserStore:            userStore,
		CrawlerConfigStore:   crawlerConfigStore,
		UserResultsStore:     userResultsStore,
		SessionKey:           config.SessionKey,
		CrawlerWebserviceURL: config.CrawlerWebserviceURL,
	})

	log.Println("Server started")
	log.Fatal(server.Start(fmt.Sprintf(":%s", config.ServerPort)))
	log.Println("Server stopped")
}

func readConfig() *config {
	conf := &config{
		Database: new(tools.MongoConfig),
	}

	readFileConfig(conf)
	readEnvConfig(conf)
	validateConfig(conf)

	return conf
}

func readFileConfig(config *config) {
	// Get server config
	file, err := ioutil.ReadFile(configFile)

	// return if no config files
	if err != nil {
		return
	}

	if err = json.Unmarshal(file, config); err != nil {
		log.Fatal(err)
	}
}

func readEnvConfig(config *config) {
	val := *serverPort
	if len(val) > 0 {
		config.ServerPort = val
	}
	// DB
	val = *dbUser
	if len(val) > 0 {
		config.Database.URL = val
	}
	val = *dbUser
	if len(val) > 0 {
		config.Database.User = val
	}
	val = *dbPassword
	if len(val) > 0 {
		config.Database.Password = val
	}
	val = *dbName
	if len(val) > 0 {
		config.Database.Name = val
	}
	// AES
	val = *aesSecret
	if len(val) > 0 {
		config.AESSecretKey = val
	}
	val = *sessionKey
	if len(val) > 0 {
		config.SessionKey = val
	}
	// Crawler webservice
	val = *crawlerServiceURL
	if len(val) > 0 {
		config.CrawlerWebserviceURL = val
	}
}

func validateConfig(config *config) {
	// TODO: actually validate the config.
	// for now it will just get printed.
	log.Printf("AES: %v", config.AESSecretKey)
	log.Printf("db: %+v", config.Database)
	log.Printf("server port: %v", config.ServerPort)
	log.Printf("crawler webservice url: %v", config.CrawlerWebserviceURL)
	log.Printf("session key: %v", config.SessionKey)
}
