package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jspc-bots/bottom"
	"github.com/lrstanley/girc"
	"github.com/olekukonko/tablewriter"
	"github.com/robfig/cron/v3"
)

type AllowListMiddleware struct {
	allowList []string
}

func (a AllowListMiddleware) Do(ctx bottom.Context, _ girc.Event) error {
	sender := ctx["sender"].(string)

	if !contains(a.allowList, sender) {
		return fmt.Errorf("sender %s is not in the scheduler allow list", sender)
	}

	return nil
}

type Bot struct {
	bottom bottom.Bottom
	cron   *cron.Cron
}

func New(user, password, server, allows string, verify bool, c *cron.Cron) (b Bot, err error) {
	b.cron = c
	b.bottom, err = bottom.New(user, password, server, verify)
	if err != nil {
		return
	}

	b.bottom.Client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		c.Cmd.Join(Chan)
	})

	router := bottom.NewRouter()
	router.AddRoute(`schedule\s+\"(.*)\"\s+([A-Z]+)\s+\"(.*)\"`, b.addSchedule)
	router.AddRoute(`show\s+schedule[s]?`, b.showSchedule)
	router.AddRoute(`unschedule\s+(\d+)`, b.deleteSchedule)

	b.bottom.Middlewares.Push(AllowListMiddleware{strings.Split(allows, ",")})
	b.bottom.Middlewares.Push(router)

	return
}

func (b *Bot) addSchedule(originator, channel string, groups []string) (err error) {
	schedule := groups[1]
	command := groups[2]
	args := groups[3]

	target := Chan

	c := Command{
		Schedule: schedule,
		Command:  command,
		Target:   target,
		Args:     args,
		irc:      b.bottom.Client,
	}

	id, err := b.cron.AddJob(schedule, c)
	if err != nil {
		return
	}

	b.bottom.Client.Cmd.Messagef(channel, "Added new job (ID: %d) to schedule on behalf of %s", id, originator)

	return
}

func (b *Bot) deleteSchedule(originator, channel string, groups []string) (err error) {
	id, err := strconv.Atoi(groups[1])
	if err != nil {
		return
	}

	b.cron.Remove(cron.EntryID(id))

	b.bottom.Client.Cmd.Messagef(channel, "Removed job %d from schedule for %s", id, originator)

	return
}

func (b *Bot) showSchedule(_, channel string, _ []string) (err error) {
	sb := strings.Builder{}

	table := tablewriter.NewWriter(&sb)
	table.SetHeader([]string{"ID", "Schedule", "Command", "Target", "Args", "Next Run"})

	for _, entry := range b.cron.Entries() {
		c, ok := entry.Job.(Command)
		if !ok {
			continue
		}

		args := c.Args
		if len(args) > 15 {
			args = fmt.Sprintf("%s...", args[:12])
		}

		table.Append([]string{fmt.Sprintf("%v", entry.ID), c.Schedule, c.Command, c.Target, args, entry.Next.In(TZ).String()})
	}

	table.Render()

	for _, line := range strings.Split(sb.String(), "\n") {
		b.bottom.Client.Cmd.Message(channel, line)
	}

	b.bottom.Client.Cmd.Messagef(channel, "(as far as I know, it's now %s)", time.Now().In(TZ).String())

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
