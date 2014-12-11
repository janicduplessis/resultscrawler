package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/janicduplessis/resultscrawler/pkg/crypto"
	"github.com/janicduplessis/resultscrawler/pkg/store/mongo"
	"github.com/janicduplessis/resultscrawler/pkg/tools"
	"github.com/janicduplessis/resultscrawler/pkg/webserver"
)

const (
	configFile = "webserver.config.json"
)

var (
	envPort = flag.String("port", "8080", "Webserver port")
)

type config struct {
	ServerPort   string
	Database     *tools.MongoConfig
	AESSecretKey string // 16 bytes
	SessionKey   string
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
		userStore,
		crawlerConfigStore,
		userResultsStore,
		config.SessionKey,
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
	val := os.Getenv("CRAWLERSERVER_SERVER_PORT")
	if len(val) > 0 {
		config.ServerPort = val
	}
	// DB
	val = os.Getenv("CRAWLERSERVER_DB_HOST")
	if len(val) > 0 {
		config.Database.Host = val
	}
	val = os.Getenv("CRAWLERSERVER_DB_PORT")
	if len(val) > 0 {
		config.Database.Port = val
	}
	val = os.Getenv("CRAWLERSERVER_DB_USER")
	if len(val) > 0 {
		config.Database.User = val
	}
	val = os.Getenv("CRAWLERSERVER_DB_PASSWORD")
	if len(val) > 0 {
		config.Database.Password = val
	}
	val = os.Getenv("CRAWLERSERVER_DB_NAME")
	if len(val) > 0 {
		config.Database.Name = val
	}
	// AES
	val = os.Getenv("CRAWLERSERVER_AES_SECRET_KEY")
	if len(val) > 0 {
		config.AESSecretKey = val
	}
	val = os.Getenv("CRAWLERSERVER_SESSION_KEY")
	if len(val) > 0 {
		config.SessionKey = val
	}
}
