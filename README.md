# IRC Timer Bot

The IRC bot does two things:

1. It runs a series of pre-determined irc commands on a schedule
1. It runs some user-supplied, ephemeral irc commands on a schedule

These commands are useful for a number of reasons:

1. We can update topics regularly
1. We can send daily messages, like news headlines, or daily calendar updates

Things to consider:

1. Predetermined commands are hardcoded to a file
1. Commands passed via messages in IRC are ephemeral: we don't persist them. If there's a need, then PRs are welcome: I just don't want to have to make decisions about how these things persist on day one


## Predetermined commands

Consider the file `schedule.toml`, which is determined by the env var `SCHEDULE_TOML`

```toml
[set-topic]
schedule: "@midnight"
command: "TOPIC"
target: "#my-chan"
args: "Welcome to my channel. The date is {{ .Date }}"
```
