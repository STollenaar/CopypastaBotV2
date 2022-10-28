package speakCommand

import (
	"bytes"
	"context"
	"copypastabot/lib/markovCommand"
	"copypastabot/util"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/bwmarrin/discordgo"
	"github.com/jfreymuth/oggvorbis"
	"github.com/jonas747/dca"
)

var (
	pollyClient *polly.Client

	queue  []*Queue
	stream *dca.StreamingSession
	Bot    *discordgo.Session
)

type Queue struct {
	synthed *polly.SynthesizeSpeechOutput
	guildID string
	userID  string
}

func init() {
	// node, _ = link.AddNode(context.TODO(), lavalink.NodeConfig{
	// 	Name:        "copypastabot", // a unique node name
	// 	Host:        "localhost",
	// 	Port:        "2333",
	// 	Password:    "",
	// 	Secure:      false, // ws or wss
	// 	ResumingKey: "",    // only needed if you want to resume a lavalink session
	// })

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("personal"))
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	pollyClient = polly.NewFromConfig(cfg)

}

// Command create a tts experience for the generated markov
func Command(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if Bot == nil {
		Bot = bot
	}

	userID := interaction.Member.User.ID
	_, err := findUserVoiceState(userID)

	if err != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You must be in a VoiceChannel in order to do this!",
			},
		})
		return
	} else {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "I am preparing my beautifull voice",
			},
		})
	}

	var markov string

	parsedArguments := util.ParseArguments([]string{"url", "user", "redditpost"}, interaction)

	if url, ok := parsedArguments["Url"]; ok {
		markov, err = markovCommand.GetMarkovURL(url)
	} else if user, ok := parsedArguments["User"]; ok {
		markov, err = markovCommand.GetUserMarkov(interaction.GuildID, user)
	} else if post, ok := parsedArguments["Redditpost"]; ok {
		postCommnents := util.GetRedditPost(post)
		markov = postCommnents.Post.Body

		if postCommnents.Post.Body == "" {
			markov = postCommnents.Post.URL
		}

	} else {
		NO_VALID_ARGUMENT := "No valid argument was provided"
		Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &NO_VALID_ARGUMENT,
		})
		return
	}
	if err != nil {
		ERROR_CALLED := "Something went wrong trying to create the Markov!!"
		Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &ERROR_CALLED,
		})
		return
	}

	contents := util.BreakContent(markov, 2950)
	for _, content := range contents {
		synthed, err := pollyClient.SynthesizeSpeech(context.TODO(), &polly.SynthesizeSpeechInput{
			Text:         aws.String(content),
			TextType:     types.TextTypeText,
			OutputFormat: types.OutputFormatOggVorbis,
			Engine:       types.EngineNeural,
			VoiceId:      types.VoiceIdMatthew,
			LanguageCode: types.LanguageCodeEnUs,
		})

		if err == nil {
			queue = append(queue, &Queue{
				synthed: synthed,
				userID:  userID,
				guildID: interaction.GuildID,
			})
		}
	}
	if stream != nil {
		if finished, _ := stream.Finished(); !finished {
			return
		}
	}

	doSpeech()
}

func doSpeech() {
	if len(queue) == 0 {
		return
	}
	currentSpeech := queue[0]
	queue = queue[1:]

	vs, err := findUserVoiceState(currentSpeech.userID)
	if err != nil {
		if len(queue) > 0 {
			doSpeech()
		}
		return
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.

	// Send the buffer data.
	stream := currentSpeech.synthed.AudioStream
	r, err := oggvorbis.NewReader(stream)
	// handle error
	if err != nil {
		// handle error
		fmt.Println(err)
	}

	buffer := make([]float32, 8192)
	vc, _ := Bot.ChannelVoiceJoin(currentSpeech.guildID, vs.ChannelID, false, false)
	vc.Speaking(true)
	for {
		n, err := r.Read(buffer)

		buffed := buffer[:n]
		buff := float32ToByte(buffed)
		bbuff, _ := readOpus(buff)

		for _, buf := range bbuff {
			vc.OpusSend <- buf
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			// handle error
			fmt.Println(err)
		}

	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	if len(queue) > 0 {
		doSpeech()
	}
}

func findUserVoiceState(userid string) (*discordgo.VoiceState, error) {
	for _, guild := range Bot.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userid {
				return vs, nil
			}
		}
	}
	return nil, errors.New("could not find user's voice state")
}

// Reads an opus packet to send over the vc.OpusSend channel
func readOpus(buffed []byte) (buffer [][]byte, err error) {
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(buffed); i += 960 {
		end := i + 960

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(buffed) {
			end = len(buffed)
		}

		buffer = append(buffer, buffed[i:end])
	}

	return buffer, err
}

func float32ToByte(f []float32) []byte {
	var buf bytes.Buffer
	for _, value := range f {
		err := binary.Write(&buf, binary.BigEndian, value)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}
	}
	return buf.Bytes()
}
