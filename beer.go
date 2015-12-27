package main

import (
	"appengine"
	"encoding/json"
	"fmt"
	"os"
	"appengine/urlfetch"
)

type UntappdConfig struct {
	ClientId string
	ClientSecret string
}

type Beer struct {
	Name        string
	Brewer		string
	Description string
	Style       string
	Image		string
}

func init() {
	untappdConfig = loadUntappdConfig()
}

var (
	untappdConfig *UntappdConfig
)

func loadUntappdConfig() *UntappdConfig {
	untappd, err := os.Open("untappd.config")
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

func queryUntappd(query string, c appengine.Context) []Beer {
	type UntappedBeer struct {
		Beer_name string `json:"beer_name"`
		Beer_label string `json:"beer_label"`
		Beer_style string `json:"beer_style"`
		Beer_description string `json:"beer_description"`
	}
	type UntappdBrewery struct {
		Brewery_name string `json:"brewery_name"`
	}
	type UntappdItems struct {
		Beer UntappedBeer `json:"beer"`
		Brewery UntappdBrewery `json:"brewery"`
	}
	type UntappdBeers struct {
		Items []UntappdItems `json:"items"`
	}
	type UntappdResponse struct {
		Beers UntappdBeers `json:"beers"`
	}
	type UntappedResponseObject struct {
		Response UntappdResponse `json:"response"`
	}

	if untappdConfig == nil {
		return []Beer{}
	}

	url := fmt.Sprintf("https://api.untappd.com/v4/search/beer?client_id=%s&client_secret=%s&q=%s", untappdConfig.ClientId, untappdConfig.ClientSecret, query)

	client := urlfetch.Client(c)
	resp, err := client.Get(url)
	if err != nil {
		c.Infof(err.Error())
		return []Beer{}
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	untappdResponse := UntappedResponseObject{}
	err = decoder.Decode(&untappdResponse)

	if err != nil {
		c.Infof(err.Error())
		return []Beer{}
	}

	beers := []Beer{}

	for _, item := range untappdResponse.Response.Beers.Items {
		beers = append(beers, Beer {
					Name : item.Beer.Beer_name,
					Brewer : item.Brewery.Brewery_name,
					Description : item.Beer.Beer_description,
					Style : item.Beer.Beer_style,
					Image : item.Beer.Beer_label, 
			})
	}

	return beers
}