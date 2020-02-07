package bot

import (
	"github.com/nlopes/slack"
	"github.com/rs/zerolog/log"
)

type Bot struct {
	apiKey   string
	api      *slack.Client
	handlers []MessageHandler
}

func New(apiKey string) *Bot {
	api := slack.New(
		apiKey,
	)
	bot := &Bot{
		apiKey: apiKey,
		api:    api,
	}
	return bot
}

func (b *Bot) Handler(h MessageHandler) {
	b.handlers = append(b.handlers, h)
}

func (b *Bot) Run() {
	rtm := b.api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			log.Info().Str("event", "connected").Msgf("%v", ev.Info)

		case *slack.MessageEvent:
			log.Info().Str("event", "message").Msgf("%v", ev.Text)
			for _, h := range b.handlers {
				err := h(b.api, ev)
				if err != nil {
					log.Error().Err(err).Msg("failed to process message")
				}
			}
		case *slack.PresenceChangeEvent:
			log.Info().Str("event", "presence").Msgf("%v", ev)

		case *slack.LatencyReport:
			log.Info().Str("event", "latency").Msgf("%v", ev.Value)

		case *slack.RTMError:
			log.Error().Err(ev).Msg("rtm error")

		case *slack.InvalidAuthEvent:
			log.Error().Msg("invalid credentials")
			return

		default:
			log.Debug().Msgf("%v", msg.Data)
		}
	}
}

type MessageHandler func(*slack.Client, *slack.MessageEvent) error
