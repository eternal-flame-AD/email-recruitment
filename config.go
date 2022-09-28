package recruitment

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/jinzhu/configor"
)

var Config AppConfig
var Prospects []Prospect

var flagConfigFile = flag.String("c", "config.yml", "config file")
var flagProspectFile = flag.String("d", "prospects.json", "prospects file")

type AppConfig struct {
	Recruiter map[string]Recruiter
	SMTP      struct {
		From     string
		Host     string
		Port     int
		Username string
		Password string
	}
}

func LoadConfigAndData() {
	cfg := configor.New(&configor.Config{
		ErrorOnUnmatchedKeys: true,
		ENVPrefix:            "RECRUITMENT",
	})
	if err := cfg.Load(&Config, *flagConfigFile); err != nil {
		log.Panicf("Failed to load config file: %v", err)
	}
	prosF, err := os.Open(*flagProspectFile)
	if err != nil {
		log.Panicf("Failed to open prospects file: %v", err)
	}
	defer prosF.Close()
	if err := json.NewDecoder(prosF).Decode(&Prospects); err != nil {
		log.Panicf("Failed to decode prospects file: %v", err)
	}
}

func SaveProspects() {
	prosF, err := os.Create(*flagProspectFile)
	if err != nil {
		log.Panicf("Failed to create prospects file: %v", err)
	}
	defer prosF.Close()
	e := json.NewEncoder(prosF)
	e.SetIndent("", "\t")
	if e.Encode(&Prospects); err != nil {
		log.Panicf("Failed to encode prospects file: %v", err)
	}
}
