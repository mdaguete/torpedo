package main

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/nlopes/slack"

	"torpedobot/giphy"
	"torpedobot/common"
	"torpedobot/multibot"
)

func GiphyProcessMessage(api *slack.Client, event *slack.MessageEvent, bot *multibot.TorpedoBot) {
	var message string
	var params slack.PostMessageParameters
	var giphyResponse giphy.GiphyResponse

	_, command, message := common.GetRequestedFeature(event.Text)
	if command != "" {
		query := url.QueryEscape(command)
		result, err := common.GetURLBytes(fmt.Sprintf("http://api.giphy.com/v1/gifs/search?q=%s&api_key=dc6zaTOxFJmzC", query))
		if err != nil {
			fmt.Printf("Get Giphy URL failed with %+v", err)
			return
		}
		err = json.Unmarshal(result, &giphyResponse)
		if err != nil {
			fmt.Printf("Error unmarshalling Giphy: %+v", err)
			return
		}
		if giphyResponse.Meta.Status == 200 {
			attachment := slack.Attachment{
				Color:     "#36a64f",
				Title:     command,
				TitleLink: giphyResponse.Data[0].URL,
				ImageURL:  giphyResponse.Data[0].Images.OriginalImage.URL,
			}
			params.Attachments = []slack.Attachment{attachment}
			message = ""
		} else {
			message = "Your request to Giphy could not be processed"
		}
	}
	bot.PostMessage(event.Channel, message, api, params)
}
