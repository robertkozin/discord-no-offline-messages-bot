package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	onlineUsers       = make(map[string]struct{})
	usersLastNotified = make(map[string]time.Time)
)

func main() {
	if err := run(); err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
	}
}

func run() (err error) {
	dg, _ := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	dg.Identify.LargeThreshold = 250
	dg.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentGuildPresences | discordgo.IntentGuilds

	dg.SyncEvents = false
	dg.StateEnabled = false

	dg.AddHandler(ready)
	dg.AddHandler(presenceUpdate)
	dg.AddHandler(guildCreate)
	dg.AddHandler(messageCreate)

	if err = dg.Open(); err != nil {
		return fmt.Errorf("err starting discordgo: %w", err)
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sig

	if err = dg.Close(); err != nil {
		return fmt.Errorf("err closing discordgo: %w", err)
	}

	return
}

func ready(s *discordgo.Session, e *discordgo.Ready) {
	for _, guild := range e.Guilds {
		for _, presence := range guild.Presences {
			updateOnlineUsers(presence.User.ID, presence.Status)
		}
	}
}

func presenceUpdate(s *discordgo.Session, e *discordgo.PresenceUpdate) {
	updateOnlineUsers(e.User.ID, e.Presence.Status)
}

func guildCreate(s *discordgo.Session, e *discordgo.GuildCreate) {
	for _, presence := range e.Presences {
		updateOnlineUsers(presence.User.ID, presence.Status)
	}
}

func updateOnlineUsers(userID string, status discordgo.Status) {
	if status == discordgo.StatusOffline {
		delete(onlineUsers, userID)
	} else {
		onlineUsers[userID] = struct{}{}
	}
}

func messageCreate(s *discordgo.Session, e *discordgo.MessageCreate) {
	var (
		isOnline     bool
		now          = time.Now()
		lastNotified time.Time
		err          error
	)

	if e.Author.Bot {
		return
	}

	_, isOnline = onlineUsers[e.Author.ID]
	if isOnline {
		return
	}

	lastNotified, _ = usersLastNotified[e.Author.ID]
	if now.Sub(lastNotified) < time.Hour {
		return
	}

	_, err = s.ChannelMessageSendReply(
		e.ChannelID,
		"How are you using Discord? You're offline.",
		&discordgo.MessageReference{
			MessageID: e.ID,
			ChannelID: e.ChannelID,
			GuildID:   e.GuildID,
		})
	if err != nil {
		fmt.Println("err replying to message:", err)
		return
	}

	usersLastNotified[e.Author.ID] = now
}
