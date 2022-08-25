package commands

import (
	"context"
	"copypastabot/util"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/dgolink"
	"github.com/disgoorg/disgolink/lavalink"
	"github.com/jonas747/dca"
	"github.com/nint8835/parsley"
)

var pollyClient *polly.Client

var queue []*Queue
var stream *dca.StreamingSession
var link *dgolink.Link
var node lavalink.Node

type Queue struct {
	synthed *polly.SynthesizeSpeechOutput
	message *discordgo.MessageCreate
}

func SpeakInit(parser *parsley.Parser) {
	link = dgolink.New(Bot.Session)
	node, _ = link.AddNode(context.TODO(), lavalink.NodeConfig{
		Name:        "copypastabot", // a unique node name
		Host:        "localhost",
		Port:        "2333",
		Password:    "",
		Secure:      false, // ws or wss
		ResumingKey: "",    // only needed if you want to resume a lavalink session
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("personal"))
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	pollyClient = polly.NewFromConfig(cfg)

	parser.NewCommand("speak", "Speak the generated markov", SpeakCommand)
}

// SpeakCommand create a tts experience for the generated markov
func SpeakCommand(message *discordgo.MessageCreate, args commandArgs) {
	_, err := Bot.findUserVoiceState(message.Author.ID)

	if err != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintln("You must be in a VoiceChannel in order to do this!"))
	}

	var markov string
	if _, err := url.ParseRequestURI(args.Word); err == nil {
		markov, err = MarkovURLCommand(message, args, false)
	} else if strings.Contains(args.Word, "@") {
		markov, err = MarkovUserCommand(message, args, false)
	} else if args.Word != "" {
		postCommnents := util.GetRedditPost(args.Word)
		markov = postCommnents.Post.Body

		if postCommnents.Post.Body == "" {
			markov = postCommnents.Post.URL
		}

	} else {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintln("Not a argument was provided"))
		return
	}
	if err != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintln("Something went wrong"))
		fmt.Println(err)
		return
	}
	contents := util.BreakContent(markov, 2950)
	for _, content := range contents {
		synthed, err := pollyClient.SynthesizeSpeech(context.TODO(), &polly.SynthesizeSpeechInput{
			Text:         aws.String(content),
			TextType:     types.TextTypeText,
			OutputFormat: types.OutputFormatMp3,
			Engine:       types.EngineNeural,
			VoiceId:      types.VoiceIdMatthew,
			LanguageCode: types.LanguageCodeEnUs,
		})

		if err == nil {
			queue = append(queue, &Queue{
				synthed: synthed,
				message: message,
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
	var buffer = make([][]byte, 0)

	vs, err := Bot.findUserVoiceState(currentSpeech.message.Author.ID)
	if err != nil {
		if len(queue) > 0 {
			doSpeech()
		}
		return
	}

	vc, _ := Bot.ChannelVoiceJoin(currentSpeech.message.GuildID, vs.ChannelID, false, false)

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
	stream := currentSpeech.synthed.AudioStream

	outFile, err := os.Create("tts.mp3")
	// handle err
	if err != nil {
		fmt.Println(err)
		return
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, stream)
	// handle err
	if err != nil {
		fmt.Println(err)
		return
	}

	buffer, err = readOpus(stream)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Send the buffer data.
	for _, buff := range buffer {
		vc.OpusSend <- buff
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	if len(queue) > 0 {
		doSpeech()
	}
}

func (s DiscordBot) findUserVoiceState(userid string) (*discordgo.VoiceState, error) {
	for _, guild := range s.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userid {
				return vs, nil
			}
		}
	}
	return nil, errors.New("could not find user's voice state")
}

// Reads an opus packet to send over the vc.OpusSend channel
func readOpus(source io.ReadCloser) (buffer [][]byte, err error) {
	var opuslen int16
	for {
		err = binary.Read(source, binary.LittleEndian, &opuslen)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				err = source.Close()
				if err != nil {
					fmt.Println(err)
					return
				}
				break
			}
			fmt.Println(err)
			return nil, err
		}

		if opuslen < 0 {
			break
		}
		var opusframe = make([]byte, opuslen)
		err = binary.Read(source, binary.LittleEndian, &opusframe)

		if err != nil {
			fmt.Println(err)
			return
		}

		buffer = append(buffer, opusframe)
	}

	return buffer, nil
}
