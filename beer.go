package main

import (
	"encoding/json"
	// "fmt"
	// "net/http"
	"os"
	// "net/url"
)

type UntappdConfig struct {
	ClientId string
	ClientSecret string
}

type Beer struct {
	Name        string
	Description string
	Style       string
}

func init() {
	untappdConfig = loadUntappdConfig()
}

var (
	untappdConfig *UntappdConfig
)

func loadUntappdConfig() *UntappdConfig {
	untappd, err := os.Open("untappd.json")
	if err != nil {
		return nil
	}
	decoder := json.NewDecoder(untappd)
	config := UntappdConfig{}
	err = decoder.Decode(&config)
	if err != nil {
		return nil
	}
	return &config
}

func queryUntappd(query string) []Beer {
	if untappdConfig == nil {
		return []Beer{}
	}

	// url := fmt.Sprintf("https://api.untappd.com/v4/search/beer?q=Pliny?client_id=%s&client_secret=%s&q=", untappdConfig.ClientId, untappdConfig.ClientSecret, url.QueryEscape(query))

	// resp, err := http.Get(url)
	// if err != nil {
	// 	return []Beer{}
	// }

	return []Beer{}	
}