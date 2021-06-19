# IRC Timer Bot

The IRC bot runs a series of pre-determined irc commands on a schedule. It also accepts new commands from trusted users. These users need to be passed to the bot via an env var, and the bot restarted because:

1. The permissions the bot potentially needs (like chanops for setting topics) means the bot could be easily abused
1. Restarting the bot is a conscious decision, and so is less likely to be done accidentally/ through security flaw

These commands are useful for a number of reasons:

1. We can update topics regularly
1. We can send daily messages, like news headlines, or daily calendar updates

Things to consider:

1. Predetermined commands are hardcoded to a file
1. Ephemeral commands (as in: added via irc) are not persisted: they will go away when the bot restarts.
1. Changes/ new commands require the bot to be restarted

## Configuration

Configuration comes from the environment:

1. `SCHEDULE_TOML` - see below; a toml file containing pre-determined commands
1. `SASL_USER` and `SASL_PASSWORD` - username/password combo for your IRC server
1. `SERVER` - IRC server to connect to, in `irc://servr:port` / `ircs://server:port` form
1. `VERIFY_TLS` - whether, on TLS enabled servers, to verify certs; useful for self-signed localhost server
1. `TZ` - timezone to apply schedules to, in `Asia/Seoul` / `Europe/London` form; if empty, loads UTC
1. `ALLOW_LIST` - comma seperated list of nicks which are allowed to update schedules


## Predetermined commands

Consider the file `schedule.toml`, which is determined by the env var `SCHEDULE_TOML`

```toml
[set-topic]
schedule: "@midnight"
command: "TOPIC"
target: "#my-chan"
args: "Welcome to my channel. The date is {{ .Date }}"
```
