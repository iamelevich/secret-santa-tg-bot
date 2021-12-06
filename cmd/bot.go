package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"secret-santa-go-bot/pkg/updates"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/dig"
)

func initMongo(ctx context.Context) (*mongo.Client, error) {
	timoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return mongo.Connect(timoutCtx, options.Client().
		ApplyURI(os.Getenv("MONGODB_URL")))
}

func initBot() *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_API_TOKEN"))
	if err != nil {
		log.Panic().Err(err).Stack().Msg("No bot token")
	}
	bot.Debug = true

	log.Info().Msg(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))
	return bot
}

func wrapError(e error) {
	if e != nil {
		log.Panic().Err(e).Stack()
	}
}

func main() {
	args := os.Args[1:]
	envFile := ".env"
	if len(args) > 0 {
		envFile = args[0]
	}
	err := godotenv.Load(envFile)
	if err != nil {
		log.Error().Str("env_file", envFile).Msg("Error loading .env file. Starting without it")
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	ctx, cancelFunc := context.WithCancel(context.Background())
	c := dig.New()
	wrapError(c.Provide(func() context.Context {
		return ctx
	}))
	wrapError(c.Provide(initBot))
	wrapError(c.Provide(initMongo))

	_ = c.Invoke(func(ctx context.Context, bot *tgbotapi.BotAPI, client *mongo.Client) {
		defer func() {
			cancelFunc()
			bot.StopReceivingUpdates()
			if err = client.Disconnect(ctx); err != nil {
				log.Error().Err(err).Stack().Msg("Mongo disconnect error")
			}
			log.Info().Msg("Finally finished!")
		}()

		go updates.CheckUpdates(ctx, bot, client)

		sigs := make(chan os.Signal, 2)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM) // subscribe to system quit
		<-sigs
	})
}
