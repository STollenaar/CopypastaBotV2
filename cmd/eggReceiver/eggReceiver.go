package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"github.com/common-nighthawk/go-figure"
	"github.com/stollenaar/copypastabotv2/internal/util"
)

var (
	//go:embed alphabet.flf
	font string

	channelID string
	regN      = regexp.MustCompile("[0-9]")
	regS      = regexp.MustCompile("[a-zA-Z?]")
	reader    = strings.NewReader(font)
	date      time.Time
)

type Event struct {
	Message string `json:"message"`
}

func main() {
	lambda.Start(handler)
	// handler(context.TODO())
}

func init() {
	c, err := util.ConfigFile.GetDiscordChannelID()
	if err != nil {
		log.Fatal(err)
	}
	channelID = c

	d, err := time.Parse("2006-01-02", util.ConfigFile.DATE_STRING)
	if err != nil {
		log.Fatal(err)
	}
	date = d
}

func handler(ctx context.Context, event *Event) error {

	var output []string
	if event != nil && event.Message != "" {
		input := strings.Split(event.Message, " ")
		for _, in := range input {
			tmp := figure.NewFigureWithFont(in, reader, true).String()
			tmp = equalise(tmp)
			tmp = regS.ReplaceAllString(tmp, ":egg:")
			tmp = strings.ReplaceAll(tmp, " ", "||<:jam:1258555405701873776>||")
			output = append(output, tmp)
			reader = strings.NewReader(font)
		}
	} else {

		until := int(math.Ceil(time.Until(date).Hours() / 24))

		input := []string{strconv.Itoa(until)}
		if util.ConfigFile.DATE_STRING == "0000-01-01" {
			input[0] = "EGG?"

			for _, in := range input {
				tmp := figure.NewFigureWithFont(in, reader, true).String()
				tmp = equalise(tmp)
				tmp = regS.ReplaceAllString(tmp, ":egg:")
				tmp = strings.ReplaceAll(tmp, " ", "||<:jam:1258555405701873776>||")
				output = append(output, tmp)
				reader = strings.NewReader(font)
			}

		} else if until < 0 {
			return nil
		} else if until == 0 {
			input[0] = "EGG"
			input = append(input, "IS")
			input = append(input, "DONE")
			for _, in := range input {
				tmp := figure.NewFigureWithFont(in, reader, true).String()
				tmp = equalise(tmp)
				tmp = regS.ReplaceAllString(tmp, ":chicken:")
				tmp = strings.ReplaceAll(tmp, " ", "||<:jar:>||")
				output = append(output, tmp)
				reader = strings.NewReader(font)
			}

		} else {
			tmp := figure.NewFigureWithFont(input[0], reader, true).String()
			tmp = equalise(tmp)
			tmp = regN.ReplaceAllString(tmp, ":egg:")
			tmp = strings.ReplaceAll(tmp, " ", "||<:jam:1258555405701873776>||")
			output = append(output, tmp)
		}
	}
	var embeds []*discordgo.MessageEmbed
	for _, in := range output {
		embeds = append(embeds, &discordgo.MessageEmbed{
			Description: in,
			Type:        discordgo.EmbedTypeRich,
		})
	}
	// embeds[0].Title = "EGG STATUS"

	response := discordgo.MessageCreate{
		Message: &discordgo.Message{
			Embeds:  embeds,
			Content: "",
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
		return err
	}

	resp, err := util.SendRequest("POST", channelID, "", util.MESSAGE, data)
	if resp != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		bodyData := buf.String()

		bodyString := string(bodyData)
		fmt.Println(resp, bodyString)
	}

	return err
}

func equalise(input string) string {
	rows := strings.Split(input, "\n")

	var max int
	for _, row := range rows {
		chars := len(strings.Split(row, ""))
		if chars > max {
			max = chars
		}
	}
	for i, row := range rows {
		if row == "" {
			continue
		}
		for len(strings.Split(row, "")) < max {
			row += " "
		}
		rows[i] = row
	}

	return strings.Join(rows, "\n")
}
