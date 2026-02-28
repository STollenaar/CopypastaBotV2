package speak

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jonas747/dca"
	"github.com/stollenaar/aws-rotating-credentials-provider/credentials/filecreds"
	"github.com/stollenaar/copypastabotv2/internal/commands/chat"
	"github.com/stollenaar/copypastabotv2/internal/commands/markov"
	"github.com/stollenaar/copypastabotv2/internal/database"
	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
)

type Queue struct {
	synthed     *polly.SynthesizeSpeechOutput
	sqsObject   util.Object
	userID      string
	dbID        int64
	isLastChunk bool
}

type SpeakCommand struct {
	Name        string
	Description string
}

var (
	pollyClient *polly.Client

	mu          sync.Mutex
	queues      map[string][]*Queue    // per-guild queue
	guildActive map[string]bool        // per-guild "playback in progress" flag
	guildVC     map[string]*voice.Conn // per-guild active voice connection

	SpeakCmd = SpeakCommand{
		Name:        "speak",
		Description: "Listen to the beauty of the bot",
	}
)

func init() {
	if os.Getenv("AWS_SHARED_CREDENTIALS_FILE") != "" {
		provider := filecreds.NewFilecredentialsProvider(os.Getenv("AWS_SHARED_CREDENTIALS_FILE"))
		pollyClient = polly.New(polly.Options{
			Credentials: provider,
			Region:      os.Getenv("AWS_REGION"),
		})
	} else {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			panic("configuration error, " + err.Error())
		}
		pollyClient = polly.NewFromConfig(cfg)
	}

	queues = make(map[string][]*Queue)
	guildActive = make(map[string]bool)
	guildVC = make(map[string]*voice.Conn)
}

func (s SpeakCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	err := event.DeferCreateMessage(util.ConfigFile.SetEphemeral() == discord.MessageFlagEphemeral)
	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}

	sub := event.SlashCommandInteractionData()

	if len(sub.Options) == 0 {
		_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Content: util.Pointer("You must provide at least 1 argument"),
		})
		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err))
		}
		return
	}

	speakObject := util.Object{
		Token:         event.Token(),
		Command:       event.Data.CommandName(),
		GuildID:       event.GuildID().String(),
		ApplicationID: event.ApplicationID().String(),
	}

	if user, ok := sub.Options["user"]; ok {
		speakObject.Type = "user"
		speakObject.Data = user.Snowflake().String()
		resp, err := util.ConfigFile.SendStatsBotRequest(speakObject)
		speakObject.Data = resp.Data
		if err != nil {
			slog.Error("Encountered an error while processing the speak command", slog.Any("err", err))
			return
		}
	} else if redditPost, ok := sub.Options["redditpost"]; ok {
		speakObject.Data = redditPost.String()
		speakObject.Type = "redditpost"
	} else if url, ok := sub.Options["url"]; ok {
		speakObject.Data = markov.HandleURL(url.String())
		speakObject.Type = "url"
	} else if message, ok := sub.Options["chat"]; ok {
		speakObject.Type = "chat"
		data, err := chat.GetChatGPTResponse("speak", message.String(), event.User().ID.String())
		if err != nil {
			slog.Error("Error interacting with chatgpt", slog.Any("err", err))
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Content: util.Pointer("If you see this, and error likely happened. Whoops"),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err))
			}
			return
		}
		speakObject.Data = data.Response
	}

	synthData(speakObject, event.User().ID.String(), event.Client())
}

func (s SpeakCommand) CreateCommandArguments() []discord.ApplicationCommandOption {
	return []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "url",
			Description: "URL of the page to make a markov chain from",
			Required:    false,
		},
		discord.ApplicationCommandOptionUser{
			Name:        "user",
			Description: "User to create a markov chain of",
			Required:    false,
		},
		discord.ApplicationCommandOptionString{
			Name:        "redditpost",
			Description: "Reddit post ID",
			Required:    false,
		},
		discord.ApplicationCommandOptionString{
			Name:        "chat",
			Description: "Chat message for the copypastabot",
			Required:    false,
		},
	}
}

// synthData synthesizes all content chunks via Polly, appends them to the
// per-guild queue, and starts playback if the guild is not already active.
func synthData(object util.Object, userID string, bot *bot.Client) {
	if object.Type == "redditpost" {
		post := util.GetRedditPost(object.Data)
		object.Data = post.Post.Body
	}

	dbID, err := database.EnqueueSpeakItem(database.QueueRecord{
		GuildID: object.GuildID,
		UserID:  userID,
		Content: object.Data,
		CmdType: object.Type,
		CmdName: object.Command,
		AppID:   object.ApplicationID,
		Token:   object.Token,
	})
	if err != nil {
		slog.Error("Error persisting speak item to DB", slog.Any("err", err))
	}

	contents := util.BreakContent(object.Data, 2950)
	for i, content := range contents {
		resp, err := util.WrapIntoSSML(content, "system")
		textType := types.TextTypeText
		engine := types.EngineNeural
		if err != nil {
			slog.Error("Error wrapping into SSML", slog.Any("err", err))
		} else {
			content = resp.Choices[0].Message.Content
			// if !strings.Contains(content, "<speak>") {
			// 	content = "<speak>" + content + "</speak>"
			// }
			textType = types.TextTypeSsml
			engine = types.EngineStandard
		}
		slog.Debug("Content to be synthesized", slog.String("content", content))
		synthed, err := pollyClient.SynthesizeSpeech(context.TODO(), &polly.SynthesizeSpeechInput{
			Text:         aws.String(content),
			TextType:     textType,
			OutputFormat: types.OutputFormatMp3,
			Engine:       engine,
			VoiceId:      types.VoiceIdMatthew,
			LanguageCode: types.LanguageCodeEnUs,
		})
		if err != nil {
			slog.Error("Error synthesizing content", slog.Any("err", err))
			continue
		}

		mu.Lock()
		queues[object.GuildID] = append(queues[object.GuildID], &Queue{
			synthed:     synthed,
			userID:      userID,
			sqsObject:   object,
			dbID:        dbID,
			isLastChunk: i == len(contents)-1,
		})
		mu.Unlock()
	}

	// Atomically check-and-set the active flag so only one goroutine starts doSpeech per guild.
	mu.Lock()
	shouldStart := !guildActive[object.GuildID]
	if shouldStart {
		guildActive[object.GuildID] = true
	}
	mu.Unlock()

	if shouldStart {
		doSpeech(object.GuildID, bot)
	}
}

func doSpeech(guildID string, bot *bot.Client) {
	mu.Lock()
	if len(queues[guildID]) == 0 {
		guildActive[guildID] = false
		mu.Unlock()
		slog.Debug("Queue is empty", slog.String("guild_id", guildID))
		return
	}
	currentSpeech := queues[guildID][0]
	queues[guildID] = queues[guildID][1:]
	mu.Unlock()

	d := "Playing now"
	response := discord.MessageUpdate{
		Content: &d,
	}
	data, _ := json.Marshal(response)
	if _, snerr := strconv.Atoi(currentSpeech.sqsObject.Token); snerr != nil {
		_, err := util.SendRequest("PATCH", currentSpeech.sqsObject.ApplicationID, currentSpeech.sqsObject.Token, util.WEBHOOK, data)
		if err != nil {
			slog.Error("Error sending patch", slog.Any("err", err))
		}
		defer func() {
			_, err := util.SendRequest("DELETE", currentSpeech.sqsObject.ApplicationID, currentSpeech.sqsObject.Token, util.WEBHOOK, data)
			if err != nil {
				slog.Error("Error sending delete", slog.Any("err", err))
			}
		}()
	}

	vs, found := bot.Caches.VoiceState(snowflake.MustParse(guildID), snowflake.MustParse(currentSpeech.userID))
	if !found {
		// User left the VC; skip this item and try the next one.
		doSpeech(guildID, bot)
		return
	}

	// Sleep for a specified amount of time before playing the sound.
	time.Sleep(250 * time.Millisecond)

	stream := currentSpeech.synthed.AudioStream
	conn := bot.VoiceManager.CreateConn(snowflake.MustParse(guildID))

	err := conn.Open(context.TODO(), *vs.ChannelID, false, false)

	// Stay connected between items; disconnect only when the queue is drained.
	mu.Lock()
	hasMore := len(queues[guildID]) > 0
	mu.Unlock()
	if !hasMore {
		defer conn.Close(context.TODO())
	}

	if err != nil {
		slog.Error("Error joining voice channel", slog.Any("err", err))
	}

	mu.Lock()
	guildVC[guildID] = &conn
	mu.Unlock()

	err = conn.SetSpeaking(context.TODO(), voice.SpeakingFlagMicrophone)
	if err != nil {
		slog.Error("Error setting speaking to true", slog.Any("err", err))
	}

	outFile, err := os.CreateTemp("/tmp", "*.mp3")
	if err != nil {
		slog.Error("Error creating temp mp3 file", slog.Any("err", err))
	}
	defer func() {
		err := outFile.Close()
		if err != nil {
			slog.Error("Error closing tmp file", slog.Any("err", err))
		}
	}()

	_, err = io.Copy(outFile, stream)
	if err != nil {
		slog.Error("Error copying stream to temp file", slog.Any("err", err))
	}

	encodeSession, err := dca.EncodeFile(outFile.Name(), dca.StdEncodeOptions)
	if err != nil {
		slog.Error("Error DCA encoding file", slog.Any("err", err))
	}

	writeOpus(encodeSession, conn.UDP())
	// Make sure everything is cleaned up.
	err = os.Remove(outFile.Name())
	if err != nil {
		slog.Error("Error removing file", slog.Any("err", err))
	}

	if currentSpeech.isLastChunk {
		if err := database.SetSpeakItemStatus(currentSpeech.dbID, "done"); err != nil {
			slog.Error("Error marking speak item as done", slog.Any("err", err))
		}
	}

	err = conn.SetSpeaking(context.TODO(), voice.SpeakingFlagNone)
	if err != nil {
		slog.Error("Error setting speaking to false", slog.Any("err", err))
	}

	mu.Lock()
	guildVC[guildID] = nil
	mu.Unlock()

	// Sleep for a specified amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	doSpeech(guildID, bot)
}

// Reenqueue re-processes speak queue items loaded from the DB on startup.
// It marks each old record done first to prevent double-replay on the next restart.
func Reenqueue(items []database.QueueRecord, bot *bot.Client) {
	for _, item := range items {
		if err := database.SetSpeakItemStatus(item.ID, "done"); err != nil {
			slog.Error("Error marking stale speak item as done", slog.Any("err", err))
		}
		obj := util.Object{
			Type:          item.CmdType,
			Command:       item.CmdName,
			Data:          item.Content,
			GuildID:       item.GuildID,
			ApplicationID: item.AppID,
			Token:         item.Token,
		}
		synthData(obj, item.UserID, bot)
	}
}

func writeOpus(encodeSession *dca.EncodeSession, w io.Writer) {
	defer encodeSession.Cleanup()

	ticker := time.NewTicker(time.Millisecond * 20)
	defer ticker.Stop()

	var frameLen int16
	// Don't wait for the first tick, run immediately.
	for ; true; <-ticker.C {
		err := binary.Read(encodeSession, binary.LittleEndian, &frameLen)
		if err != nil {
			panic("error reading file: " + err.Error())
		}

		// Copy the frame.
		_, err = io.CopyN(w, encodeSession, int64(frameLen))
		if err != nil && err != io.EOF {
			return
		}
	}
}
