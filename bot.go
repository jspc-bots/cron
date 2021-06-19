package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lrstanley/girc"
	"github.com/olekukonko/tablewriter"
	"github.com/robfig/cron/v3"
)

type Bot struct {
	client    *girc.Client
	cron      *cron.Cron
	routing   map[*regexp.Regexp]handlerFunc
	allowList []string
}

type handlerFunc func(originator string, groups [][]byte) error

func New(user, password, server, allows string, verify bool, c *cron.Cron) (b Bot, err error) {
	b.cron = c
	b.allowList = strings.Split(allows, ",")

	u, err := url.Parse(server)
	if err != nil {
		return
	}

	config := girc.Config{
		Server: u.Hostname(),
		Port:   must(strconv.Atoi(u.Port())).(int),
		Nick:   Nick,
		User:   Nick,
		Name:   Nick,
		SASL: &girc.SASLPlain{
			User: user,
			Pass: password,
		},
		SSL: u.Scheme == "ircs",
		TLSConfig: &tls.Config{
			InsecureSkipVerify: !verify,
		},
	}

	b.client = girc.New(config)
	err = b.addHandlers()

	return
}

func (b *Bot) addHandlers() (err error) {
	b.client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		c.Cmd.Join(Chan)
	})

	b.routing = make(map[*regexp.Regexp]handlerFunc)

	// Matches `schedule "@every 10ms" PRIVMSG "hello world"` in order to add a new schedule
	b.routing[regexp.MustCompile(`schedule\s+\"(.*)\"\s+([A-Z]+)\s+\"(.*)\"`)] = b.addSchedule
	b.routing[regexp.MustCompile(`show\s+schedule[s]?`)] = b.showSchedule
	b.routing[regexp.MustCompile(`unschedule\s+\"(.+)\"`)] = b.deleteSchedule

	// Route messages
	b.client.Handlers.Add(girc.PRIVMSG, b.messageRouter)

	return
}

func (b Bot) messageRouter(c *girc.Client, e girc.Event) {
	var err error

	// skip messages older than a minute (assume it's the replayer)
	cutOff := time.Now().Add(0 - time.Minute)
	if e.Timestamp.Before(cutOff) {
		// ignore
		return
	}

	msg := []byte(e.Last())

	for r, f := range b.routing {
		if r.Match(msg) {
			err = f(e.Source.Name, r.FindAllSubmatch(msg, -1)[0])
			if err != nil {
				log.Printf("%v error: %s", f, err)
			}

			return
		}
	}

	// Ignore; not a message for us
}

func (b *Bot) addSchedule(originator string, groups [][]byte) (err error) {
	if !contains(b.allowList, originator) {
		err = fmt.Errorf("%s is not in the scheduler allow list", originator)

		b.client.Cmd.Messagef(Chan, err.Error())
	}

	schedule := string(groups[1])
	command := string(groups[2])
	args := string(groups[3])

	target := Chan

	c := Command{
		Schedule: schedule,
		Command:  command,
		Target:   target,
		Args:     args,
		irc:      b.client,
	}

	_, err = b.cron.AddJob(schedule, c)
	if err != nil {
		b.client.Cmd.Messagef(Chan, "Couldn't add to the schedule: %v", err)

		return
	}

	b.client.Cmd.Message(Chan, "Added to schedule, the schedule now looks like:")
	b.showSchedule("", make([][]byte, 0))

	return
}

func (b *Bot) deleteSchedule(originator string, groups [][]byte) (err error) {
	if !contains(b.allowList, originator) {
		err = fmt.Errorf("%s is not in the scheduler allow list", originator)

		b.client.Cmd.Messagef(Chan, err.Error())
	}

	id, err := strconv.Atoi(string(groups[1]))
	if err != nil {
		err = fmt.Errorf("%s is not a valid ID", groups[1])

		b.client.Cmd.Messagef(Chan, err.Error())

		return
	}

	b.cron.Remove(cron.EntryID(id))

	b.client.Cmd.Message(Chan, "Removed from schedule, the schedule now looks like:")
	b.showSchedule("", make([][]byte, 0))

	return
}

func (b *Bot) showSchedule(_ string, _ [][]byte) (err error) {
	sb := strings.Builder{}

	table := tablewriter.NewWriter(&sb)
	table.SetHeader([]string{"ID", "Schedule", "Command", "Target", "Args", "Next Run"})

	for _, entry := range b.cron.Entries() {
		c := entry.Job.(Command)

		table.Append([]string{fmt.Sprintf("%v", entry.ID), c.Schedule, c.Command, c.Target, c.Args, entry.Next.In(TZ).String()})
	}

	table.Render()

	for _, line := range strings.Split(sb.String(), "\n") {
		b.client.Cmd.Message(Chan, line)
	}

	b.client.Cmd.Messagef(Chan, "(as far as I know, it's now %s)", time.Now().In(TZ).String())

	return
}

func contains(l []string, s string) bool {
	for _, ss := range l {
		if s == ss {
			return true
		}
	}

	return false
}
