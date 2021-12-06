package updates

import (
	"context"
	"fmt"
	"log"

	"secret-santa-go-bot/pkg/users"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	KeyboardOptionRegistration string = "Начать регистрацию"
	KeyboardOptionMyData       string = "Мои данные"
	KeyboardOptionChangeName   string = "Изменить имя"
	KeyboardOptionChangeWish   string = "Изменить пожелания"
	KeyboardOptionGiftedInfo   string = "Кому там я дарю?"
	KeyboardOptionDrawLots     string = "Провести жеребьевку"
	KeyboardOptionTestDrawLots string = "Провести тестовую жеребьевку"
	KeyboardOptionBack         string = "Назад"
)

var states = StatesData{
	users.UserStateStart: {
		KeyboardButtons: [][]string{
			{KeyboardOptionRegistration},
		},
		IsTextAllowed: false,
		NextState:     users.UserStateRegistrationName,
	},
	users.UserStateRegistrationName: {
		KeyboardButtons: [][]string{},
		IsTextAllowed:   true,
		NextState:       users.UserStateRegistrationWish,
	},
	users.UserStateRegistrationWish: {
		KeyboardButtons: [][]string{},
		IsTextAllowed:   true,
		NextState:       users.UserStateWait,
	},
	users.UserStateChangeName: {
		KeyboardButtons: [][]string{
			{KeyboardOptionBack},
		},
		IsTextAllowed: true,
		NextState:     users.UserStateWait,
	},
	users.UserStateChangeWish: {
		KeyboardButtons: [][]string{
			{KeyboardOptionBack},
		},
		IsTextAllowed: true,
		NextState:     users.UserStateWait,
	},
	users.UserStateWait: {
		KeyboardButtons: [][]string{
			{KeyboardOptionMyData, KeyboardOptionChangeName, KeyboardOptionChangeWish},
		},
		AdminButtons:  []string{KeyboardOptionDrawLots, KeyboardOptionTestDrawLots},
		IsTextAllowed: false,
		NextState:     "",
	},
	users.UserStateComplete: {
		KeyboardButtons: [][]string{
			{KeyboardOptionGiftedInfo},
		},
		IsTextAllowed: false,
		NextState:     "",
	},
}

func CheckUpdates(
	ctx context.Context,
	bot *tgbotapi.BotAPI,
	client *mongo.Client,
) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	usersCollection := users.GetUsersCollection(client)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		ca := CommandArgs{
			Ctx:             ctx,
			Bot:             bot,
			Update:          &update,
			UsersCollection: usersCollection,
		}

		user, err := users.GetUser(ctx, usersCollection, &update)
		if err != nil {
			Send(ca, "Ошибка: "+err.Error())
			log.Panicf("Get user error: %v", err)
		}
		ca.User = &user
		ca.State = states.GetByUserState(user.State)

		if user.State == "start" && update.Message.IsCommand() && update.Message.Command() == "start" {
			CommandStart(ca)
			continue
		}

		if update.Message.Sticker != nil {
			Send(ca, fmt.Sprintf("Уважаемое %s, иди ты нахуй со своими стикерами!", user.GetName()))
			SendSticker(ca.Bot, update.Message.Chat.ID, StickerWriteInTextFileID)
			continue
		}

		if update.Message.Voice != nil {
			Send(ca, "Ты серьезно? Мне еще эти ебучие голосовые слушать?")
			continue
		}

		if update.Message.Video != nil {
			Send(ca, "Ну бля, мог бы и круглое видео записать... Никакого уважения")
			continue
		}

		if update.Message.VideoNote != nil {
			Send(ca, "Hy хоть видео круглое, и то хорошо.")
			continue
		}

		if update.Message.Document != nil {
			Send(ca, "Не страдай херней")
			continue
		}

		if update.Message.Photo != nil {
			Send(ca, "Ты хоть нюдсы скидывай, нах мне это?")
			continue
		}

		if !ca.State.IsTextAllowed && !ca.State.HasOption(ca.User, update.Message.Text) {
			ErrorText(ca)
		} else if ca.State.HasOption(ca.User, update.Message.Text) {
			switch update.Message.Text {
			case KeyboardOptionRegistration:
				CommandRegistration(ca)
			case KeyboardOptionChangeName:
				CommandChangeName(ca)
			case KeyboardOptionChangeWish:
				CommandChangeWish(ca)
			case KeyboardOptionMyData:
				CommandMyInfo(ca)
			case KeyboardOptionDrawLots:
				CommandDrawLots(ca)
			case KeyboardOptionTestDrawLots:
				CommandTestDrawLots(ca)
			case KeyboardOptionBack:
				CommandBack(ca)
			case KeyboardOptionGiftedInfo:
				CommandGiftedInfo(ca)
			default:
				Send(ca, "В процессе разработки")
			}
		} else if ca.State.IsTextAllowed && update.Message.Text != "" {
			if user.State == users.UserStateRegistrationName {
				RegistrationName(ca)
			} else if user.State == users.UserStateRegistrationWish {
				RegistrationWish(ca)
			} else if user.State == users.UserStateChangeName {
				ChangeName(ca)
			} else if user.State == users.UserStateChangeWish {
				ChangeWish(ca)
			}
		} else {
			ErrorText(ca)
		}
	}
}
