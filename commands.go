package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"

	"strings"
	"strconv"
	"math/rand"
	"time"
	"regexp"
	)

// command and reply structs
type Command struct {
	Restricted bool // restrict commands
	Triggers []string // command triggers
	Replies []*Reply // weighted replies
	Action func(*discordgo.Session, *discordgo.MessageCreate) bool // action for a bot to perform
	}

type Reply struct {
	Message string // message string the BOT will return
	Weight int // adjust how likely bot is going to use this to reply
	}
func createReply(m string, w int) *Reply {
	return &Reply {
		Message: m,
		Weight: w,
		}
	}
func (c Command) Reply() string {
	i := rand.Intn(100)

	for _, r := range c.Replies {
		if i > r.Weight {
			i -= r.Weight
			} else {
			return r.Message
			}
		}

	log.Error("Error :: failed finding a reply")
	return "h͔͇̙̺̱ͥ̒ͫ͛ͮ͞ͅͅe͚̪͚̺͔͚̒͌̾̄͗̐ͬ͞l̵̼͕̠̥̬̟̥͚̉͛̿̅ͨͣ͂̉̐͜͠p̢̱̍̋ ̥̘͍ͧ̇m̷̦͍͙̯̩̤̙̅ͮ͒ͪ̎̈͐͞ͅe̖̭̭̺͊̅ͤ̈́̃̈́͒͘"
	}

// allow users logging bugs
var DEBUG *Command = &Command {
	Restricted: false,
	Triggers: []string {
		"!bug",
		"!debug",
		},
	Replies: []*Reply{
		createReply("Ok, sorry, że zawaliłam :cry:", 33),
		createReply("Bug zapisany. ピッ", 66),
		createReply("ツーツーツー\n:bomb:", 100),
		},
	Action: actionDebugReport,
	}
func actionDebugReport(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	log.Error("BUG\n:: "+ m.Author.Username +"\n:: "+ m.Content)

	return true
	}

// allow admins changing status
var STATUS *Command = &Command {
	Restricted: true,
	Triggers: []string {
		"!status",
		"zmień status",
		"ustaw status",
		},
	Action: actionSetStatus,
	}
func actionSetStatus(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	command := strings.Split(m.Content, "|")

	status := command[1]

	command = strings.Split(command[0], " ")
	idle := 0
	i, err := strconv.Atoi(command[len(command)-1] )
	if (err == nil) && (i > 0) {
		idle = int(time.Now().Unix()) -i
		}

	s.UpdateStatus(idle, status)

	log.Info("STATUS\n:: "+ m.Author.Username +"\n:: "+ m.Content)

	return true
	}

// add a counter to memory so the shitposting will end. one day.
var LENNY *Command = &Command {
	Restricted: false,
	Triggers: []string {
		"͡° ͜ʖ ͡°",
		" ͜ʖ",
		"!shitpost",
		},
	Replies: []*Reply{
		createReply("ಠ_ಠ", 10),
		createReply("(✧ω✧)", 20),
		createReply("(⍤◞ ⍤)", 40),
		createReply("☞ ( ͡° ͜ʖ ͡°) ☞", 70),
		createReply("( ͡° ͜ʖ ͡°)", 100),
		},
	}

// rip JORDINE
var RIP *Command = &Command {
	Restricted: false,
	Triggers: []string {
		"!rip",	"rip", "r.i.p",
		"[*]",
		"jordin",
		},
	Replies: []*Reply{
		createReply("fff", 10),
		createReply("bg Kakao fix", 20),
		createReply("rip", 30),
		createReply("[*]", 50),
		},
	}

// cleanup hat
var CLEANUP *Command = &Command {
	Restricted: true,
	Triggers: []string {
		"!clean",
		"!cleanup",
		"!czyść",
		},
	Action: actionCleanChat,
	}
func actionCleanChat(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	channel := m.ChannelID

	command := strings.Split(m.Content, " ")
	count := command[len(command)-1]

	IDList := []string{}
	messageList, _ := s.ChannelMessages(channel, 0, "", "")

	i, err := strconv.Atoi(count)
	if (err == nil) && (i > 0) {
		messageList, _ = s.ChannelMessages(channel, i, m.ID, "")
		for _, msg := range messageList {
			IDList = append(IDList, msg.ID)
			}
		} else {
		// get date a month before today
		_t := time.Now()
		t := time.Date(_t.Year(), _t.Month()-1, 0, 0, 0, 0, 0, time.UTC)
		mask := strconv.Itoa(t.Year()) +"-"+ padL(strconv.Itoa(int(t.Month())+1), "0", 2)

		// find message month old
		messageList, _ = s.ChannelMessages(channel, 100, m.ID, "")
		for _, msg := range messageList {
			if has(m.Timestamp[:7], mask) {
				IDList = append(IDList, msg.ID)
				}
			}
		}

	err = s.ChannelMessagesBulkDelete(m.ChannelID, IDList)
	if err != nil {
		log.Error(err)
		return false
		}

	log.Info("CLEANUP\n:: "+ m.Author.Username)

	// clean up the command
	err = s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		log.Error(err)
		return false
		}

	return true
	}

// alerts
var ALERT *Command = &Command {
	Restricted: false,
	Triggers: []string {
		"!alert",
		"przypom",
		"pamięta",
		"informu",
		"uwaga",
		},
	Replies: []*Reply{
		createReply(":ok:", 10),
		createReply("Ok, zapisane! :bookmark:", 30),
		createReply("OK :ok_hand:", 50),
		createReply("Przypomnę.",750),
		createReply("Ok, zapisane!", 100),
		},
	Action: actionCreateAlert,
	}
func actionCreateAlert(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	// parse command for date and body
	msg := parseMessage(m.ContentWithMentionsReplaced(), s.State.Ready.User.Username)

	//rWeekday := regexp.MustCompile("(?i)(?P<weekday>poniedziałek|wtorek|środę|czwartek|piątek|sobotę|niedzielę).*?(?P<day>[0-9]{1,2})(?P<month>\\.(?P<monthN>[0-9]{1,2})| (?P<monthD>(sty|lut|mar|kwie|maj|czer|lipi|sier|wrze|paźd|list|grud)\\w+))?")

	rDate := regexp.MustCompile("(?i)(?P<date>(?P<day>[0-9]{1,2})(?P<month>\\.(?P<monthN>[0-9]{1,2})| (?P<monthS>sty\\w+|lut\\w+|mar\\w+|kwie\\w+|maj\\w+|czer\\w+|lipi\\w+|sier\\w+|wrze\\w+|paźd\\w+|list\\w+|grud\\w+)))")

	rHours := regexp.MustCompile("(?i)(?P<time>(?P<hour>[0-9]{1,2}):(?P<minutes>[0-9]{2}))")
	rChannel := regexp.MustCompile("(?i)(?P<channel>(veli\\w+|calph\\w+|balen\\w+|medi\\w+|vale\\w+) [0-9])")
	rNode := regexp.MustCompile("(?i)(?P<node>(node|noda) (([A-Z][ a-z]+)+))")

	//weekday := rWeekday.FindStringSubmatch(msg)
	date := rDate.FindStringSubmatch(msg)
	hours := rHours.FindStringSubmatch(msg)
	channel := rChannel.FindStringSubmatch(msg)
	node := rNode.FindStringSubmatch(msg)

	if len(date)==0 || len(hours)==0 || len(channel)==0 || len(node)==0 {
		member, _ := MEMORY.getMember(m.Author.ID)	

		errorMsg := "Nie mogę zapisać wydarzenia, brakuje kilku informacji:\n"
		if  len(date)==0 {errorMsg += "daty\n"}
		if  len(hours)==0 {errorMsg += "godziny wojny\n"}
		if  len(channel)==0 {errorMsg += "kanału z numerem\n"}
		if  len(node)==0 {errorMsg += "noda o który będzie walka\n"}

		s.ChannelMessageSend(member.Priv, errorMsg)

		return false
		}

	month := time.Month(0)
	if len(date[4]) > 0 {
		i, er := strconv.Atoi(date[4])
		month = time.Month(i+1)
		if er != nil {
			member, _ := MEMORY.getMember(m.Author.ID)	
			s.ChannelMessageSend(member.Priv, "Nie mogę zapisać wydarzenia, zapisz miesiąc w dacie inaczej :(")
			return false
			}
		}
	if len(date[5]) > 0 {
		month = time.Month(monthCheck(date[5])+1)
		}

	day, er := strconv.Atoi(date[2])
	if er != nil {
		member, _ := MEMORY.getMember(m.Author.ID)	
		s.ChannelMessageSend(member.Priv, "Nie mogę zapisać wydarzenia, zapisz dzień w dacie inaczej :(")
		return false
		}

	_t := time.Now()
	t := time.Date(_t.Year(), month, day, 0, 0, 0, 0, time.UTC)

	reason := "Dzisiaj o "+ hours[0] + " walczymy o "+ strings.Title(node[3]) +" na kanale "+ strings.Title(channel[0]) +".\nZbiórka pół godziny wcześniej."
	createAlert(m.Author.ID, t, reason)

	log.Info("ALERT\n:: "+ m.Author.Username +"\n:: "+ reason)

	return true
	}

// pull patchnotes
var GETUPDATES *Command = &Command {
	Restricted: false,
	Triggers: []string {
		"!patch",
		"!update",
		"!announcement",
		},
	Action: actionGetUpdates,
	}
func actionGetUpdates(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	l := "https://www.blackdesertonline.com/" +crawl("https://www.blackdesertonline.com/news/list/announcement")

	member, _ := MEMORY.getMember(m.Author.ID)
	if member == nil {
		return false
		}
	s.ChannelMessageSend(member.Priv, ":newspaper: " +l)

	log.Info("UPDATES\n:: "+ m.Author.Username +"\n:: "+ m.Content)

	return true
	}
var GETEVENTS *Command = &Command {
	Restricted: false,
	Triggers: []string {
		"!event",
		},
	Action: actionGetEvents,
	}
func actionGetEvents(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	l := "https://www.blackdesertonline.com/" +crawl("https://www.blackdesertonline.com/news/list/event")

	member, _ := MEMORY.getMember(m.Author.ID)
	if member == nil {
		return false
		}
	s.ChannelMessageSend(member.Priv, ":newspaper: " +l)

	log.Info("EVENTS\n:: "+ m.Author.Username +"\n:: "+ m.Content)

	return true
	}
var GETMAINTENANCE *Command = &Command {
	Restricted: false,
	Triggers: []string {
		"!patch",
		"!update",
		},
	Action: actionGetMaintenance,
	}
func actionGetMaintenance(s *discordgo.Session, m *discordgo.MessageCreate) bool {	
	l := "https://www.blackdesertonline.com/" +crawl("https://www.blackdesertonline.com/news/list/maintenance")

	member, _ := MEMORY.getMember(m.Author.ID)
	if member == nil {
		return false
		}
	s.ChannelMessageSend(member.Priv, ":newspaper: " +l)

	log.Info("MAINTENACE\n:: "+ m.Author.Username +"\n:: "+ m.Content)

	return true
	}

// purge the swear
var triggersSwear = []string{"kutas", "dupe", "pedal", "kurw", "qrw", "choler", "jeb", "huj", "pierd", "suki", "sra", "sry", "shit", "fuck", "piss", "bitch", "puss", "puta", "spiepr", "doopa",}
var SWEAR *Command = &Command {
	Restricted: false,
	Triggers: triggersSwear,
	Action: actionSwear,
	}
func actionSwear(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	// if they ever add editing foreign messages, replace swearword with :cookie:
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		log.Error(err)
		return false
		}

	// send a chill cookie to the offender
	_, err = s.ChannelMessageSend(m.ChannelID, "<@"+ m.Author.ID +"> :cookie:")
	if err != nil {
		log.Error(err)
		return false
		}

	log.Info("SWEAR\n:: "+ m.Author.Username)

	return true
}

var PURGE *Command = &Command {
	Restricted: true,
	Triggers: []string {
		"!swear",
		"!purge",
		"!flush",
		},
	Action: actionPurge,		
	}
func actionPurge(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	channel := m.ChannelID
	r := regexp.MustCompile("(?i)[^ ]*("+ strings.Join(triggersSwear, "|") +")[^ ,.]*")
	// get date a month before today
	IDList := []string {}

	messageList, _ := s.ChannelMessages(channel, 100, m.ID, "")
	for _, msg := range messageList {
		if len(r.FindAllString(m.Content, -1)) > 0 {
			IDList = append(IDList, msg.ID)
			}
		}

	err := s.ChannelMessagesBulkDelete(m.ChannelID, IDList)
	if err != nil {
		log.Error(err)
		return false
		}

	// clean up the command
	err = s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		log.Error(err)
		return false
		}

	log.Info("SWEAR\n:: "+ m.Author.Username)

	return true
	}

// export the fun
var COMMANDS []*Command = []*Command{
	DEBUG,
	STATUS,
	LENNY,
	RIP,
	CLEANUP,
	ALERT,
	GETUPDATES, GETEVENTS, GETMAINTENANCE,
	SWEAR,
	PURGE,
	}