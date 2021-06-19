package main

import (
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/robfig/cron/v3"
)

const (
	Nick = "scheduler"
	Chan = "#dashboard"
)

var (
	ScheduleFile = os.Getenv("SCHEDULE_TOML")
	Username     = os.Getenv("SASL_USER")
	Password     = os.Getenv("SASL_PASSWORD")
	Server       = os.Getenv("SERVER")
	VerifyTLS    = os.Getenv("VERIFY_TLS") == "true"
	AllowList    = os.Getenv("ALLOW_LIST")

	TZ = must(time.LoadLocation(os.Getenv("TZ"))).(*time.Location)
)

func must(i interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}

	return i
}

func main() {
	var commands Commands

	if ScheduleFile == "" {
		log.Print("$SCHEDULE_TOML is empty, skipping pre-defined commands")
	} else {
		f, err := os.ReadFile(os.Getenv("SCHEDULE_TOML"))
		if err != nil {
			panic(err)
		}

		if _, err := toml.Decode(string(f), &commands); err != nil {
			panic(err)
		}
	}

	c, err := New(Username, Password, Server, AllowList, VerifyTLS, cron.New(cron.WithLocation(TZ)))
	if err != nil {
		panic(err)
	}

	for _, command := range commands {
		command.irc = c.bottom.Client

		_, err = c.cron.AddJob(command.Schedule, command)
		if err != nil {
			panic(err)
		}
	}

	go func() {
		log.Panic(c.bottom.Client.Connect())
	}()

	c.cron.Run()
}
