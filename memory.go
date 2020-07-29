package main

import (
	log "github.com/Sirupsen/logrus"
    "github.com/bwmarrin/discordgo"

	"time"
	"regexp"
	)

type Memory struct {
	Guild string
	Alerts []*Alert
	Members []*Member
	Roles []*Role
	Channels []*Channel
	}

const (
	userNotWatched = iota
	userWatched = iota
	userRecruited = iota
	userRanked = iota
	)

type Alert struct {
	ID string // setter's ID
	Date time.Time // date for an alert to go off
	Reason string // stored message to remind
	}
func createAlert(i string, d time.Time, r string) *Alert {
	return &Alert {
		ID: i,
		Date: d,
		Reason: r,
		}
	}
func (MEMORY *Memory) deleteAlert(a *Alert) bool {
	for i, _a := range MEMORY.Alerts {
		if _a == a {
			MEMORY.Alerts = append(MEMORY.Alerts[:i], MEMORY.Alerts[i+1:]...)
			return true
			}
		}
	return false
	}
func checkAlerts(s *discordgo.Session, MEMORY *Memory) {
	now := time.Now()

	for _, alert := range MEMORY.Alerts {
		if now.Sub(alert.Date) <= 0 { // the alert is ready to pop
			karczma := MEMORY.getChannels("karczma")[0]
			autor, _ := MEMORY.getMember(alert.ID)

			hopki := MEMORY.getRolesID("hopki")[0]

			_, err := s.ChannelMessageSend(karczma.ID, "<@"+ hopki +"!\n"+ alert.Reason)
			if err != nil {
				log.Error(err)
				return
				}

			log.Info("ALERT (done)\n:: "+ autor.Name +"\n::"+ alert.Reason)

			MEMORY.deleteAlert(alert) // we done
			}
		}
	}

type Member struct {
	Name string
	ID string
	Status int
	Priv string
	Roles []*Role
	}
func createMember(i string, n string, s int, p string, r []*Role) *Member {
	return &Member {
		ID: i,
		Name: n,
		Status: s,
		Priv: p,
		Roles: r,
		}
	}
func (MEMORY *Memory) setMembers(s *discordgo.Session) {
	members, _ := s.GuildMembers(MEMORY.Guild, 0, 100)
	priv, _ := s.UserChannels()

	for _, m := range members {
		if m.User.ID != BOTID { // ignore self
			// memorise user's roles			
			roles := MEMORY.copyRoles(m.Roles...)

			// memorise user's priv channel
			channel := ""
			for _, ch := range priv {
				if ch.Recipient.ID == m.User.ID {
					channel = ch.ID
					break
					}
				}
			if channel == "" { // if private channel doesn't exist yet, create it
				ch, _ := s.UserChannelCreate(m.User.ID)
				channel = ch.ID
				}

			MEMORY.Members = append(MEMORY.Members, createMember(
				m.User.ID,
				m.User.Username,
				userNotWatched,
				channel,
				roles,
				))
			}
		}
	}
func (MEMORY *Memory) getMember(id string) (*Member, int) {
	for i, mm := range MEMORY.Members {
		if mm.ID == id {
			return mm, i
			}
		}
		return nil, -1
	}
func (m *Member) hasRole(name ...string) bool {
	for _, n := range name {
		for _, r := range m.Roles {
			if r.Name != n {
				return false
				}
			}
		}
	return true
	}
func (m *Member) isAdmin() bool {
	for _, r := range m.Roles {
		if r.Admin {
			return true
			}
		}
	return false
	}
func (m *Member) addRoles(MEMORY Memory, name ...string) {
	for _, n := range name {
		is := false
		for _, r := range m.Roles { // check if user has no such role
				if n == r.Name {
					is = true
					}
				}

		if !is {
			m.Roles = append(m.Roles, MEMORY.getRoles(n)...)
			}
		}
	}

func classCheck(m string) (classes []string) {
	// regexes for BDO classes
	rBerserker := regexp.MustCompile("(?P<Berserker>[a-ząę]*[sz]er[a-ząę]*k[a-ząę]*)")
	rMusa := regexp.MustCompile("(?P<Musa>mu[sz][a-ząę]*)")
	rMaehwa := regexp.MustCompile("(?P<Maehwa>m[aw]{1,2}h?w[a-ząę]*)")
	rNinja := regexp.MustCompile("(?P<Ninja>nin[a-ząę]*)")
	rKunoichi := regexp.MustCompile("(?P<Kunoichi>kun[a-ząę]*)")
	rRanger := regexp.MustCompile("(?P<Ranger>rang[a-ząę]*)")	
	rWizard := regexp.MustCompile("(?P<Wizard>wiz[a-ząę]*)")
	rWitch := regexp.MustCompile("(?P<Witch>wit?c[a-ząę]*)")
	rValkyrie := regexp.MustCompile("(?P<Valkyrie>[vw]al[a-ząę]*)")
	rDarkKnight := regexp.MustCompile("(?P<DarkKnight>(Dark[a-ząę]*)? ?(Kn[aąę]*)?)")

	if len(rBerserker.FindStringSubmatch(m))>0 {classes = append(classes, "Berserker")}
	if len(rMusa.FindStringSubmatch(m))>0 {classes = append(classes, "Musa")}
	if len(rMaehwa.FindStringSubmatch(m))>0 {classes = append(classes, "Maehwa")}
	if len(rNinja.FindStringSubmatch(m))>0 {classes = append(classes, "Ninja")}
	if len(rKunoichi.FindStringSubmatch(m))>0 {classes = append(classes, "Kunoichi")}
	if len(rRanger.FindStringSubmatch(m))>0 {classes = append(classes, "Ranger")}
	if len(rWizard.FindStringSubmatch(m))>0 {classes = append(classes, "Wizard")}
	if len(rWitch.FindStringSubmatch(m))>0 {classes = append(classes, "Witch")}
	if len(rValkyrie.FindStringSubmatch(m))>0 {classes = append(classes, "Valkyrie")}
	if len(rDarkKnight.FindStringSubmatch(m))>0 {classes = append(classes, "Dark Knight")}

	return
	}

type Role struct {
	Name string
	ID string
	Admin bool
	}
func createRole(i string, n string, b bool) *Role {
	return &Role {
		ID: i,
		Name: n,
		Admin: b,
		}
	}
func (MEMORY *Memory) setRoles(s *discordgo.Session, admin ...string) {
	roles, _ := s.GuildRoles(MEMORY.Guild)
	for _, r := range roles {
		admin := false
		if is(r.Name, "almighty", "Oficer") {
			admin = true
			}
		MEMORY.Roles = append(MEMORY.Roles, createRole(r.ID, r.Name, admin))
		}
	}
func (MEMORY *Memory) copyRoles(discordRoles ...string) (roles []*Role) {
	for _, rID := range discordRoles {
		for _, r := range MEMORY.Roles {
			if rID == r.ID {
				roles = append(roles, r)
				}
			} 
		}
	return
	}
func (MEMORY *Memory) getRoles(name ...string) (roles []*Role) {
	for _, r := range MEMORY.Roles {
		for _, n := range name {
			if r.Name == n {
				roles = append(roles, r)
				}
			}
		}
		return
	}
func (MEMORY *Memory) getRolesID(name ...string) (roles []string) {
	for _, r := range MEMORY.Roles {
		for _, n := range name {
			if r.Name == n {
				roles = append(roles, r.ID)
				}
			}
		}
		return
	}

type Channel struct {
	ID string
	Name string
	Type string
	}
func createChannel(i string, n string, t string) *Channel {
	return &Channel {
		ID: i,
		Name: n,
		Type: t,
		}
	}
func (MEMORY *Memory) setChannels(s *discordgo.Session) {
	channels, _ := s.GuildChannels(MEMORY.Guild)
	for _, ch := range channels {
		MEMORY.Channels = append(MEMORY.Channels, createChannel(ch.ID, ch.Name, ch.Type))
		}
	}
func (MEMORY *Memory) getChannels(name ...string) (channels []*Channel) {
	for _, ch := range MEMORY.Channels {
		for _, n := range name {
			if ch.Name == n {
				channels = append(channels, ch)
				}
			}
		}
		return
	}

// fire function every set hours
func poll(f func(s *discordgo.Session, MEMORY *Memory), s *discordgo.Session, MEMORY *Memory) {
	for {
		<-time.After(24 *time.Hour)
		go f(s, MEMORY)
		}
	}