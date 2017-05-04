package multibot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"

	"torpedobot/memcache"
	"crypto/md5"
	"io"
)


type TorpedoBot struct {
	caches map[string]*memcache.MemCacheType
	commandHandlers map[string]func(*slack.Client, *slack.MessageEvent, *TorpedoBot)
	config struct {
		api_keys []string
	}
}

func (tb *TorpedoBot) PostMessage(channel, message string, api *slack.Client, parameters ...slack.PostMessageParameters) {
	var params slack.PostMessageParameters

	if len(parameters) > 0 {
		params = parameters[0]
	}
	channelID, timestamp, err := api.PostMessage(channel, message, params)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}

func (tb *TorpedoBot) processChannelEvent(api *slack.Client, event *slack.MessageEvent) {
	messageTS, _ := strconv.ParseFloat(event.Timestamp, 64)
	jitter := int64(time.Now().Unix()) - int64(messageTS)

	if jitter < 10 && strings.HasPrefix(event.Text, "!") {
		command := strings.TrimPrefix(event.Text, "!")
		found := 0
		for handler := range tb.commandHandlers {
			if strings.HasPrefix(strings.Split(command, " ")[0], handler) {
				found += 1
				tb.commandHandlers[handler](api, event, tb)
				break
			}
		}
		fmt.Printf("PROCESS! -> %s", command)
		if found == 0 {
			tb.PostMessage(event.Channel, fmt.Sprintf("Could not process your message: !%s. Command unknown. Send !help for list of valid commands.", command), api)
		}
	}
}

func (tb *TorpedoBot) RunBot(apiKey string) {
	api := slack.New(apiKey)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	api.SetDebug(true)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	// TODO: Move this somewhere else
	for msg := range rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)
			// Replace #general with your Channel ID
			// rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", "#general"))

		case *slack.MessageEvent:
			fmt.Printf("Message: %v\n", ev)
			go tb.processChannelEvent(api, ev)

		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			fmt.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:
			// Ignore other events..
			//fmt.Printf("Unexpected: %v\n", msg.Data)
		}
	}

}

func (tb *TorpedoBot) RunLoop() {
	for {
		time.Sleep(time.Second)
	}
}

func (tb *TorpedoBot) RunBots() {
	for _, key := range tb.config.api_keys {
		go tb.RunBot(key)
	}
	tb.RunLoop()
}

func (tb *TorpedoBot) RegisterHandlers(handlers map[string]func(*slack.Client, *slack.MessageEvent, *TorpedoBot)) {
	tb.commandHandlers = handlers
	return
}

func (tb *TorpedoBot) GetCommandHandlers() (handlers map[string]func(*slack.Client, *slack.MessageEvent, *TorpedoBot)) {
	return tb.commandHandlers
}

func (tb *TorpedoBot) GetCreateCache(name string) (cache *memcache.MemCacheType) {
	value, success := tb.caches[name]
	if ! success {
		cache = memcache.New()
		tb.caches[name] = cache
	} else {
		cache = value
	}
	return
}

func (tb *TorpedoBot) GetCachedItem(name string) (item string){
	cache := *tb.GetCreateCache(name)
	if cache.Len() > 0 {
		fmt.Printf("\nUsing cached quote...%v\n", cache.Len())
		key := ""
		for key = range cache.Cache() {
			break
		}
		quote, _ := cache.Get(key)
		cache.Delete(key)
		item = quote
	}
	return
}


func (tb *TorpedoBot) SetCachedItems(name string, items map[int]string) (item string){
	cache := *tb.GetCreateCache(name)
	for idx := range items {
		my_hash := md5.New()
		io.WriteString(my_hash, items[idx])
		message := fmt.Sprintf("%x", my_hash.Sum(nil))
		_, ok := cache.Get(message)
		if ! ok {
			cache.Set(message, items[idx])
		}
	}

	item = items[0]
	//
	my_hash := md5.New()
	io.WriteString(my_hash, item)
	message := fmt.Sprintf("%x", my_hash.Sum(nil))
	//
	cache.Delete(message)
	return
}


func New(api_keys []string) (bot *TorpedoBot) {
	bot = &TorpedoBot{}
	bot.config.api_keys = api_keys
	bot.caches = make(map[string]*memcache.MemCacheType)
	return
}
