package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/stollenaar/copypastabotv2/internal/util"
	statsUtil "github.com/stollenaar/statisticsbot/util"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

var (
	pollyClient *polly.Client

	queue  []*Queue
	stream *dca.StreamingSession
	bot    *discordgo.Session

	guildVC map[string]*discordgo.VoiceConnection
)

type Queue struct {
	synthed   *polly.SynthesizeSpeechOutput
	sqsObject statsUtil.SQSObject
	userID    string
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	pollyClient = polly.NewFromConfig(cfg)
	guildVC = make(map[string]*discordgo.VoiceConnection)
	token, err := util.ConfigFile.GetDiscordToken()
	if err != nil {
		log.Fatalf("error fetching discord token: %v", err)
	}

	bot, err = discordgo.New("Bot " + token)
	if err != nil {
		bot.Close()
		log.Fatalf("error creating new discord session: %v", err)
	}

	err = bot.Open()
	if err != nil {
		bot.Close()
		log.Fatalf("error creating session websocket: %v", err)
	}
}

func main() {
	defer bot.Close()
	lambda.Start(handler)
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	var sqsObject statsUtil.SQSObject

	err := json.Unmarshal([]byte(sqsEvent.Records[0].Body), &sqsObject)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(sqsObject)
	return synthData(sqsObject)
}

// Command create a tts experience for the generated markov
func synthData(object statsUtil.SQSObject) error {
	resp, err := util.SendRequest("GET", object.ApplicationID, object.Token, util.WEBHOOK, []byte{})
	var bodyString string
	if resp != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		bodyData := buf.String()

		bodyString = string(bodyData)
		fmt.Println(resp, bodyString)
	}
	if err != nil {
		return fmt.Errorf("error sending request %v", err)
	}
	if object.Type == "redditpost" {
		post := util.GetRedditPost(object.Data)
		object.Data = post.Post.Body
	}

	var message discordgo.Message
	err = json.Unmarshal([]byte(bodyString), &message)
	if err != nil {
		return fmt.Errorf("error parsing into interaction with data: %s, and error: %v", bodyString, err)
	}

	contents := util.BreakContent(object.Data, 2950)
	for _, content := range contents {
		synthed, err := pollyClient.SynthesizeSpeech(context.TODO(), &polly.SynthesizeSpeechInput{
			Text:         aws.String(content),
			TextType:     types.TextTypeText,
			OutputFormat: types.OutputFormatMp3,
			Engine:       types.EngineNeural,
			VoiceId:      types.VoiceIdMatthew,
			LanguageCode: types.LanguageCodeEnUs,
		})

		if err != nil {
			fmt.Printf("error synthesizing content: %v", err)
			continue
		}
		queue = append(queue, &Queue{
			synthed:   synthed,
			userID:    message.Interaction.User.ID,
			sqsObject: object,
		})
	}
	if stream != nil {
		if finished, _ := stream.Finished(); !finished {
			fmt.Println("stream is not finished and not null")
			return nil
		}
	}

	return doSpeech()
}

func doSpeech() error {
	if len(queue) == 0 {
		fmt.Println("Queue is empty")
		return nil
	}
	currentSpeech := queue[0]
	queue = queue[1:]

	d := "Playing now"
	response := discordgo.WebhookEdit{
		Content: &d,
	}
	data, _ := json.Marshal(response)
	util.SendRequest("PATCH", currentSpeech.sqsObject.ApplicationID, currentSpeech.sqsObject.Token, util.WEBHOOK, data)
	defer util.SendRequest("DELETE", currentSpeech.sqsObject.ApplicationID, currentSpeech.sqsObject.Token, util.WEBHOOK, data)

	vs, err := bot.State.VoiceState(currentSpeech.sqsObject.GuildID, currentSpeech.userID)
	if err != nil {
		return fmt.Errorf("error finding voice state: %v", err)
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.

	// Send the buffer data.
	stream := currentSpeech.synthed.AudioStream

	vc, err := bot.ChannelVoiceJoin(currentSpeech.sqsObject.GuildID, vs.ChannelID, false, false)
	if len(queue) == 0 {
		defer vc.Disconnect()
	}

	if err != nil {
		return fmt.Errorf("error joining voice channel: %v", err)
	}
	guildVC[currentSpeech.sqsObject.GuildID] = vc
	err = vc.Speaking(true)
	if err != nil {
		return fmt.Errorf("error setting speaking to true: %v", err)
	}
	// write the whole body at once
	outFile, err := os.Create("/tmp/tmp.mp3")
	if err != nil {
		return fmt.Errorf("error creating tmp mp3 file: %v", err)
	}
	defer outFile.Close()
	// handle err
	_, err = io.Copy(outFile, stream)
	if err != nil {
		return fmt.Errorf("error copying stream to tmp file: %v", err)
	}
	// Encoding a file and saving it to disk
	encodeSession, err := dca.EncodeFile("/tmp/tmp.mp3", dca.StdEncodeOptions)
	if err != nil {
		return fmt.Errorf("error dca encoding file: %v", err)
	}

	done := make(chan error)
	dca.NewStream(encodeSession, vc, done)
	err = <-done
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return fmt.Errorf("stream error: %v", err)
	}

	// Make sure everything is cleaned up, that for example the encoding process if any issues happened isnt lingering around
	encodeSession.Cleanup()
	os.Remove("/tmp/tmp.mp3")

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	if len(queue) > 0 {
		return doSpeech()
	} else {
		guildVC[currentSpeech.sqsObject.GuildID] = nil
	}
	return nil
}

// func VCInterupt(bot *discordgo.Session) {

// 	ticker := time.NewTicker(time.Duration(10) * time.Minute)

// 	go func() {
// 		for {
// 			<-ticker.C
// 			// do stuff
// 			guildChannels := checkVCs()
// 			for guild, channelVS := range guildChannels {
// 				if channelVS == nil || len(channelVS.users) == 0 {
// 					continue
// 				}
// 				user := getRandomUser(channelVS)
// 				user = strings.ReplaceAll(user, "<", "")
// 				user = strings.ReplaceAll(user, ">", "")
// 				user = strings.ReplaceAll(user, "@", "")

// 				markov, err := markovCommand.GetUserMarkov(guild, user)
// 				if err != nil {
// 					fmt.Println(err)
// 					continue
// 				}
// 				contents := util.BreakContent(markov, 2950)
// 				for _, content := range contents {
// 					synthed, err := pollyClient.SynthesizeSpeech(context.TODO(), &polly.SynthesizeSpeechInput{
// 						Text:         aws.String(content),
// 						TextType:     types.TextTypeText,
// 						OutputFormat: types.OutputFormatMp3,
// 						Engine:       types.EngineNeural,
// 						VoiceId:      types.VoiceIdMatthew,
// 						LanguageCode: types.LanguageCodeEnUs,
// 					})

// 					if err == nil {
// 						queue = append(queue, &Queue{
// 							synthed: synthed,
// 							userID:  user,
// 							guildID: guild,
// 						})
// 					}
// 				}
// 				if stream != nil {
// 					if finished, _ := stream.Finished(); !finished {
// 						return
// 					}
// 				}

// 				doSpeech()
// 			}
// 		}
// 	}()
// }

// type channelVS struct {
// 	channelID string
// 	users     []string
// }

// func getRandomUser(channelVS *channelVS) (user string) {
// 	index := rand.Intn(len(channelVS.users))
// 	return channelVS.users[index]
// }

// func checkVCs() (guildChannelStates map[string]*channelVS) {
// 	guildChannelStates = make(map[string]*channelVS)
// 	for _, guild := range Bot.State.Guilds {
// 		guild, _ = Bot.State.Guild(guild.ID)

// 		if guildVC[guild.ID] != nil {
// 			continue
// 		}
// 		channelStates := make(map[string]*channelVS)
// 		for _, vs := range guild.VoiceStates {
// 			if vs.UserID == Bot.State.User.ID {
// 				continue
// 			}

// 			if channelStates[vs.GuildID] == nil {
// 				channelStates[vs.GuildID] = &channelVS{
// 					users:     []string{vs.UserID},
// 					channelID: vs.ChannelID,
// 				}
// 			} else {
// 				channelStates[vs.GuildID].users = append(channelStates[vs.GuildID].users, vs.UserID)
// 			}
// 		}
// 		guildChannelStates[guild.ID] = getMostChannelUsers(channelStates)
// 	}
// 	return guildChannelStates
// }

// func getMostChannelUsers(channels map[string]*channelVS) (channelVS *channelVS) {
// 	for _, v := range channels {
// 		if channelVS == nil || len(v.users) > len(channelVS.users) {
// 			channelVS = v
// 		}
// 	}
// 	return channelVS
// }
