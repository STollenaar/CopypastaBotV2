package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/stollenaar/copypastabotv2/internal/util"
	statsUtil "github.com/stollenaar/statisticsbot/util"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/bwmarrin/discordgo"
)

var (
	sqsClient *sqs.Client
	Bot       *discordgo.Session
)

type channelVS struct {
	channelID string
	users     []string
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}
	sqsClient = sqs.NewFromConfig(cfg)

	token, err := util.ConfigFile.GetDiscordToken()
	if err != nil {
		log.Fatalf("error fetching discord token: %v", err)
	}

	Bot, err = discordgo.New("Bot " + token)
	if err != nil {
		Bot.Close()
		log.Fatalf("error creating new discord session: %v", err)
	}

	err = Bot.Open()
	if err != nil {
		Bot.Close()
		log.Fatalf("error creating session websocket: %v", err)
	}
}

func main() {
	defer Bot.Close()
	lambda.Start(handler)
}

func handler() error {
	guildChannels := checkVCs()
	if len(guildChannels) == 0 {
		fmt.Println("No guildchannels active")
	}

	for guild, channelVS := range guildChannels {
		if debugGuild := os.Getenv("DEBUG_GUILD"); debugGuild != "" && guild != debugGuild {
			continue
		}
		if channelVS == nil || len(channelVS.users) == 0 {
			fmt.Printf("No active voice channels for %s\n", guild)
			continue
		}
		user := getRandomUser(channelVS)
		user = strings.ReplaceAll(user, "<", "")
		user = strings.ReplaceAll(user, ">", "")
		user = strings.ReplaceAll(user, "@", "")
		sqsMessage := statsUtil.SQSObject{
			Token:         user,
			Command:       "speak",
			Type:          "user",
			Data:          user,
			GuildID:       guild,
			ApplicationID: Bot.State.User.ID,
		}
		sqsMessageData, _ := json.Marshal(sqsMessage)
		fmt.Printf("Sending sqsData: %s\n", string(sqsMessageData))
		_, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
			MessageBody: aws.String(string(sqsMessageData)),
			QueueUrl:    &util.ConfigFile.AWS_SQS_URL,
		})
		if err != nil {
			return fmt.Errorf("error sending interrupt event to sqs: %v", err)
		}
	}

	return nil
}

func getRandomUser(channelVS *channelVS) (user string) {
	index := rand.Intn(len(channelVS.users))
	return channelVS.users[index]
}

func checkVCs() (guildChannelStates map[string]*channelVS) {
	guildChannelStates = make(map[string]*channelVS)
	guildVC := make(map[string]*discordgo.VoiceConnection)
	for _, guild := range Bot.State.Guilds {
		guild, _ = Bot.State.Guild(guild.ID)

		if guildVC[guild.ID] != nil {
			continue
		}
		channelStates := make(map[string]*channelVS)
		for _, vs := range guild.VoiceStates {
			if vs.UserID == Bot.State.User.ID {
				continue
			}

			if channelStates[vs.GuildID] == nil {
				channelStates[vs.GuildID] = &channelVS{
					users:     []string{vs.UserID},
					channelID: vs.ChannelID,
				}
			} else {
				channelStates[vs.GuildID].users = append(channelStates[vs.GuildID].users, vs.UserID)
			}
		}
		guildChannelStates[guild.ID] = getMostChannelUsers(channelStates)
	}
	return guildChannelStates
}

func getMostChannelUsers(channels map[string]*channelVS) (channelVS *channelVS) {
	for _, v := range channels {
		if channelVS == nil || len(v.users) > len(channelVS.users) {
			channelVS = v
		}
	}
	return channelVS
}
