package main

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

type CityTraffic struct {
	City         string
	TrafficLevel string
	Players      int
}

type ServerTraffic struct {
	ServerName    string
	BusiestCities []CityTraffic
}

func scrape() {
	fName := "traffic.json"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()

	// Instantiate default collector
	c := colly.NewCollector()

	serversTraffic := make([]ServerTraffic, 0, 200)

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("Scrapping", r.URL.String())
	})

	// Extract details of the server
	c.OnHTML(`ul.list-group-flush`, func(e *colly.HTMLElement) {
		title := e.ChildText("h5 a") + " " + e.ChildText("h5 span")

		if title == " " {
			return
		}

		busiestCities := make([]CityTraffic, 0, 200)
		serverTraffic := ServerTraffic{
			ServerName:    title,
			BusiestCities: busiestCities,
		}

		e.ForEach(`div[id^="server_traffic"]`, func(_ int, el *colly.HTMLElement) {
			re := regexp.MustCompile(`(.*)- (.*) \(([0-9]+)\)`)
			match := re.FindStringSubmatch(el.Text)

			if match == nil {
				return
			}

			players, _ := strconv.Atoi(strings.TrimSpace(match[3]))

			cityTraffic := CityTraffic{
				City:         strings.TrimSpace(match[1]),
				TrafficLevel: strings.TrimSpace(match[2]),
				Players:      players,
			}

			serverTraffic.BusiestCities = append(serverTraffic.BusiestCities, cityTraffic)
		})

		serversTraffic = append(serversTraffic, serverTraffic)
	})

	c.Visit("https://traffic.krashnz.com")

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	// Dump json to the standard output
	enc.Encode(serversTraffic)
}
