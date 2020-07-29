package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"

	"os"
	"os/signal"
	"flag"
	"strconv"
	"strings"
	"math/rand"
	"time"	
	)

var (
	// new discordgo session
	discord *discordgo.Session

	// Memory block
	MEMORY *Memory = &Memory {
		Guild: "",
		Alerts: []*Alert {},
		Members: []*Member {},
		Roles: []*Role {},
		Channels: []*Channel {},
		}

	// Bot add url 
	// redacted ;)

	// Bot Owner
	OWNER string = "BOT_OWNER"

	// Bot Token
	TOKEN string = "BOT_TOKEN"

	// Bot ID
	BOTID string = "BOT_STRING"

	// Kotonko version
	VERSION string = "v2.0a"
	)

func onReady(s *discordgo.Session, e *discordgo.Ready) {
	// run memory watcher
	go poll(checkAlerts, s, MEMORY)

	s.UpdateStatus(0, "Black Desert Online")
	}

// when joining a new Guild
func onGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	if g.Guild.Unavailable != nil {
		return
		}
	// memorise guild
	MEMORY.Guild = g.Guild.ID
	log.Info("Joined guild "+ g.Guild.Name +" :: "+ g.Guild.ID)

	// memorise roles
	MEMORY.setRoles(s, "almighty", "Oficer")

	// memorise guild channels
	MEMORY.setChannels(s)

	// memorise guild members
	MEMORY.setMembers(s)
	}

func onGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	log.Info("GuildMemberAdd ", m.User.Username)

	// welcome user
	for _, ch := range MEMORY.Channels {
		if (ch.Name == "poczekalnia") && (ch.Type == "text") {
			s.ChannelMessageSend(ch.ID, "Cześć <@"+ m.User.ID +">! Witaj na serwerze Gildii Hopeless.")
			break
			}
		}

	// memorise user
	ch, _ := s.UserChannelCreate(m.User.ID)

	roles := MEMORY.copyRoles(m.Roles...)

	MEMORY.Members = append(MEMORY.Members, createMember(
		m.User.ID,
		m.User.Username,
		userWatched,
		ch.ID,
		roles,
		))

	// greet @user on priv
	s.ChannelMessageSend(ch.ID, "Cześć, <@"+ m.User.ID +">!")

	// grant @rekrut rank
	err := s.GuildMemberEdit(MEMORY.Guild, m.User.ID, MEMORY.getRolesID("Rekrut"))
	if err != nil {
		log.Error("Can't edit user. ", err)
		}
	}
func onGuildMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	if m.User.ID != BOTID { // ignore self
		return
		}
	log.Info("GuildMemberUpdate ", m.User.Username)

	roles := MEMORY.getRoles("Rekrut", "Hopki")
	rekrut, hopki := roles[0].ID, roles[1].ID
	member, _ := MEMORY.getMember(m.User.ID)

	if (member.Status == userWatched) && is(rekrut , m.Roles...){ // user is being watched & received rank @rekruit
		for _, mm := range MEMORY.Members {
			// priv @rekruter about new user
			if mm.hasRole("Rekrutant") {
				s.ChannelMessageSend(mm.Priv, ":grey_exclamation: nowy rekrut <@"+ m.User.ID +"> na przyjęcie do gildii.")
				}

			// priv @rekrut
			if mm.ID == m.User.ID {
				s.ChannelMessageSend(mm.Priv, "Fajnie, że chcesz do nas dołączyć! Rekruterzy już zostali powiadomieni :slight_smile:")
				s.ChannelMessageSend(mm.Priv, "Przygotuj mikrofon i dołącz do kanału głosowego #Rekrutacja.")
				}
			}
		}
	if (member.Status == userWatched) && !is(rekrut, m.Roles...) && is(hopki, m.Roles...) { // user is being watched and has been promoted from @rekrut to @hopki
		s.ChannelMessageSend(member.Priv, "To już prawie wszystko!")
		s.ChannelMessageSend(member.Priv, "Jeśli chcesz, żeby reszta wiedziała jakimi klasami grasz, to mi je napisz :smile:")

		member.Status = userRecruited
		}
	}
func onGuildMemberRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	if m.User.ID == BOTID { // ignore self
		return
		}
	member, i := MEMORY.getMember(m.User.ID)

	// remove user from memory
	MEMORY.Members = append(MEMORY.Members[:i], MEMORY.Members[:i+1]...)
	}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if (len(m.Content) < 1) || (m.Author.ID == BOTID) { //  ignore messages without mentions or properly formatted commands, or made by the BOT itself
		return
		}

	// get channel to work with
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil || channel == nil {
		log.WithFields(log.Fields{
			"channel": m.ChannelID,
			"message": m.ID,
			}).Warning("Failed to grab channel", err)
		return
		}

	// get user for admin, status check
	member, i := MEMORY.getMember(m.Author.ID)
	if i == -1 {
		log.Error("Can't pull member "+ m.Author.Username +" from MEMORY.")
		return
		}

	// prep msg
	msg := parseMessage(m.ContentWithMentionsReplaced(), s.State.Ready.User.Username)

	// parse msg for command triggers
	wasCommand := false
	for _, cmd := range COMMANDS {
		if has(strings.ToLower(msg), cmd.Triggers...) && (!cmd.Restricted || member.isAdmin()) {
			if cmd.Action != nil {
				if cmd.Action(s, m) && cmd.Replies != nil {
					go s.ChannelMessageSend(m.ChannelID, cmd.Reply())
					}

				wasCommand = true
			}
			return
			}
		}

	// parsing failed, if on priv, assume chat
	if !wasCommand && m.ChannelID == member.Priv {
		// if user is recruited
		if member.Status == userRecruited {
			classes := classCheck(msg)
			member.addRoles(*MEMORY, classes...)

			guildMember, err := s.GuildMember(MEMORY.Guild, member.ID)
			if err != nil || guildMember == nil {
				log.Error("Can't get user. ", err)
				return
				}

			roles := []string{}
			for _, r := range guildMember.Roles {
				roles = append(roles, r)
				}

			err = s.GuildMemberEdit(MEMORY.Guild, m.Author.ID, join(roles, classes...))
			if err != nil {
				log.Error("Can't edit user. ", err)
				}

			// thank user
			s.ChannelMessageSend(member.Priv, "Dzięki!")

			member.Status = userRanked
			}

		}

	// understanding message failed. ignore it
	}

func main() {
	Token := flag.String("t", "", "Discord Authentication Token")
	Owner := flag.String("o", "", "Owner ID")
	BotID := flag.String("b", "", "Bot ID")
	Shard := flag.String("s", "", "Shard ID")
	ShardCount := flag.String("c", "", "Number of Shards")

	flag.Parse()

	// if bot launch flags are not set
	if *Owner != "" {
		OWNER = *Owner
		}
	if *Token != "" {
		TOKEN = *Token
		}
	if *BotID != "" {
		BOTID = *BotID
		}

	// Setup discord
	log.Info("Starting discord session...")
	discord, err := discordgo.New("Bot "+ TOKEN)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			}).Fatal("Failed to create discord session")
		return
		}

	// Client sharding info
	discord.ShardID, _ = strconv.Atoi(*Shard)
	discord.ShardCount, _ = strconv.Atoi(*ShardCount)
	if discord.ShardCount <= 0 {
		discord.ShardCount = 1
		}
	
	// add discord event handlers
	discord.AddHandler(onReady)

	discord.AddHandler(onGuildCreate)

	discord.AddHandler(onGuildMemberAdd)
	discord.AddHandler(onGuildMemberRemove)
	discord.AddHandler(onGuildMemberUpdate)

	discord.AddHandler(onMessageCreate)

	// Connect to discord
	err = discord.Open()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,	
			}).Fatal("Failed to connect to Discord using websocket")
		return
		}

	// setup random seeds
	rand.Seed(time.Now().UTC().UnixNano())

	// Kotonko is running!
	log.Info("Kotonko "+ VERSION +" is ready! 起きています！")

	// Handling interrups
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c

	// gracefully exit
	err = discord.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,	
			}).Fatal("Failed to close Discord session")
		}
	}