package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var token string
var buffer = make([][]byte, 0)

func startBot() {
	if token == "" {
		fmt.Println("No token provided. Please run: main -t <bot token>")
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Traffic Update is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!traffic"
	if m.Content == "!traffic" {
		var discordMessage string

		serversTraffic := readJSON()

		if serversTraffic == nil {
			return
		}

		for i := 0; i < len(serversTraffic); i++ {
			serverTraffic := serversTraffic[i]

			discordMessage += fmt.Sprintf("**__%s__**\n", serverTraffic.ServerName)
			discordMessage += "**Busiest cities:**\n"

			for j := 0; j < len(serverTraffic.BusiestCities); j++ {
				cityTraffic := serverTraffic.BusiestCities[j]

				discordMessage += fmt.Sprintf(
					"- %s - %s - %d\n",
					cityTraffic.City,
					cityTraffic.TrafficLevel,
					cityTraffic.Players,
				)
			}

			discordMessage += "\n"
		}

		s.ChannelMessageSend(m.ChannelID, discordMessage)
	}
}

func readJSON() []ServerTraffic {
	fName := "traffic.json"

	fileStats, _ := os.Stat(fName)
	lastModified := fileStats.ModTime().Unix()

	if lastModified < time.Now().Add(-time.Minute*5).Unix() {
		scrape()
	}

	file, err := os.Open(fName)

	if err != nil {
		log.Fatalf("Cannot open file %q: %s\n", fName, err)
		return nil
	}

	byteValue, _ := ioutil.ReadAll(file)

	var serversTraffic []ServerTraffic

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &serversTraffic)

	defer file.Close()

	return serversTraffic
}
