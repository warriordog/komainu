package bot

import (
	"context"
	"komainu/commands"
	"komainu/storage"
	"log"
	"os"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

// Connect connects to Discord
func Connect(cfg *storage.Configuration, kvs storage.KeyValueStore) *state.State {
	var token = os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatalln("No BOT_TOKEN found in environment variables.")
	}

	state := state.New("Bot " + token)

	state.AddHandler(func(e *gateway.MessageCreateEvent) {
		if e.GuildID == 0 {
			return // It's either a private message, or an ephemeral-response command. Doesn't count.
		}

		if err := storage.See(kvs, e.GuildID, e.Author.ID); err != nil {
			log.Printf("Seen in %d: %d sent a message in %s, BUT WAS NOT RECORDED:%s\n", e.GuildID, e.Author.ID, e.ChannelID, err)
		} else {
			log.Printf("Seen in %d: %d sent a message in %s\n", e.GuildID, e.Author.ID, e.ChannelID)
		}
	})

	state.AddHandler(func(e *gateway.MessageReactionAddEvent) {
		log.Printf("Reaction in %d: %d reacted to message %s in %s with %s", e.GuildID, e.UserID, e.MessageID, e.ChannelID, e.Emoji)
	})

	commands.AddDeleteHandler(state, kvs)
	commands.AddCommandHandler(state, kvs)

	state.AddHandler(func(e *gateway.GuildCreateEvent) {
		commands.RegisterCommands(state, e.ID)
	})

	state.AddIntents(gateway.IntentGuilds |
		gateway.IntentGuildMembers |
		gateway.IntentGuildBans |
		gateway.IntentGuildEmojis |
		gateway.IntentGuildIntegrations |
		gateway.IntentGuildInvites |
		gateway.IntentGuildMessages |
		gateway.IntentGuildMessageReactions |
		gateway.IntentGuildMessageTyping |
		gateway.IntentDirectMessages |
		gateway.IntentDirectMessageReactions |
		gateway.IntentDirectMessageTyping)

	if err := state.Open(context.Background()); err != nil {
		log.Fatalln("Failed to connect to Discord:", err)
	}

	user, err := state.Me()
	if err != nil {
		log.Fatalln("Failed to get myself:", err)
	}
	log.Printf("Connected to Discord as %s#%s\n", user.Username, user.Discriminator)

	go storage.StartClosingExpiredVotes(state, kvs)

	return state
}
