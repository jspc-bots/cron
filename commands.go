package main

import (
	"log"
	"strings"
	"text/template"
	"time"

	"github.com/lrstanley/girc"
)

type TemplateValues struct {
	Time time.Time
	Date string
}

type Commands map[string]Command
type Command struct {
	Schedule string
	Command  string
	Target   string
	Args     string

	irc *girc.Client
}

func (c *Command) Event() (e *girc.Event, err error) {
	tmpl, err := template.New("").Parse(c.Args)
	if err != nil {
		return
	}

	sb := strings.Builder{}

	err = tmpl.Execute(&sb, TemplateValues{Time: time.Now().In(TZ), Date: time.Now().In(TZ).Format("2006. 01. 02")})
	if err != nil {
		return
	}

	e = &girc.Event{
		Command: c.Command,
		Params:  []string{c.Target, sb.String()},
	}

	return
}

func (c Command) Run() {
	e, err := c.Event()
	if err != nil {
		log.Print(err.Error())
		c.irc.Cmd.Messagef(Chan, "scheduled task errored :/ %v", err)

		return
	}

	c.irc.Send(e)
}
