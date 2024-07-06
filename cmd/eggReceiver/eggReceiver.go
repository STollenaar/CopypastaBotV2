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
	reg       = regexp.MustCompile("[0-9]")
	reader    = strings.NewReader(font)
	date      time.Time
)

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

func handler(ctx context.Context) error {
	until := int(math.Ceil(time.Until(date).Hours() / 24))

	var number string
	input := strconv.Itoa(until)
	if until < 0 {
		return nil
	} else if until == 0 {
		input = "EGG IS COOKED"
		number = figure.NewFigureWithFont(input, reader, true).String()
		number = equalise(number)
		number = reg.ReplaceAllString(number, ":chicken:")
		number = strings.ReplaceAll(number, " ", "||<:jar:>||")
	} else {
		number = figure.NewFigureWithFont(input, reader, true).String()
		number = equalise(number)
		number = reg.ReplaceAllString(number, ":egg:")
		number = strings.ReplaceAll(number, " ", "||<:jam:1258555405701873776>||")
	}

	response := discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content: number,
			Flags:   discordgo.MessageFlagsUrgent,
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
