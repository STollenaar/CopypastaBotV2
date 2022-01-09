package lib

import (
	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"

	commands "copypastabot/lib/commads"
)

// Bot main reference to the bot
var Bot *discordgo.Session

func Init(bot *discordgo.Session, parser *parsley.Parser) {
	Bot = bot
	commands.Init(bot, parser)
}
