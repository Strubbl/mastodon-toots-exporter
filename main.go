package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/mattn/go-mastodon"
)

var configPtr = flag.String("config", "config.json", "path to config.json")
var allStatusesPtr = flag.String("statuses", "statuses.json", "path to json for statuses")
var dummyConfigPtr = flag.Bool("dummy", false, "create a dummy config, can be used together with config flag")

// Config includes all parameters to connect to a Mastodon instance
type Config struct {
	Server           string
	ClientID         string
	ClientSecret     string
	MastodonAccount  string
	MastodonPasswort string
}

func readConfig(path string) (config *Config) {
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()
	byteConfig, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteConfig, &config)
	return config
}

func createDummyConfig(configPath string) {
	c := &Config{Server: "https://social.tchncs.de", ClientID: "132456", ClientSecret: "s0s3cr3t", MastodonAccount: "mastodon-accounts-email@address.com", MastodonPasswort: "t0pS3cr3t"}
	jsonConfig, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		log.Println("createDummyConfig json marshal error")
		log.Fatal(err)
	}
	err = ioutil.WriteFile(configPath, jsonConfig, 0644)
	if err != nil {
		log.Println("createDummyConfig write error")
		log.Fatal(err)
	}
}

func handleFlags() {
	flag.Parse()
	// log.Println("config:", *configPtr)
	// log.Println("followers:", *followersPtr)
	// log.Println("dummy config:", *dummyConfigPtr)
	if *dummyConfigPtr {
		createDummyConfig(*configPtr)
		os.Exit(0)
	}
}

func printStatuses(s []*mastodon.Status) {
	for i := range s {
		// log.Printf("\t\t@%v\t%v\t%v", s[i].ID, s[i].CreatedAt, s[i].Content)
		log.Println(s[i].ID)
		log.Println(s[i].CreatedAt.Local())
		log.Println(s[i].Content)
		log.Println(s[i].URL)
	}
}

func main() {
	// init, load config, login to mastodon
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	startTime := time.Now()
	log.Println(startTime)
	handleFlags()
	config := readConfig(*configPtr)
	c := mastodon.NewClient(&mastodon.Config{
		Server:       config.Server,
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
	})
	err := c.Authenticate(context.Background(), config.MastodonAccount, config.MastodonPasswort)
	if err != nil {
		log.Println("auth error")
		log.Fatal(err)
	}

	// get my user account
	account, err := c.GetAccountCurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// load all statuses
	var pg mastodon.Pagination
	var allStatuses []*mastodon.Status
	for {
		log.Println("Getting followers with pg.MaxID:", pg.MaxID)
		statuses, err := c.GetAccountStatuses(context.Background(), account.ID, &pg)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Number of new followers from this page:", len(statuses))
		allStatuses = append(allStatuses, statuses...)
		if pg.MaxID == "" || len(statuses) == 0 {
			break
		}
		// TODO find if toot has been exported already and break in that case
		pg.SinceID = ""
		pg.MinID = ""
		break
	}

	printStatuses(allStatuses)
	log.Println(time.Now())
}
