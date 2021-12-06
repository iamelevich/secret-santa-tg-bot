package updates

import (
	"context"
	"fmt"
	"secret-santa-go-bot/pkg/users"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommandArgs struct {
	Ctx             context.Context
	Bot             *tgbotapi.BotAPI
	Update          *tgbotapi.Update
	State           *StateData
	NextState       string
	User            *users.User
	UsersCollection *mongo.Collection
}

func Send(args CommandArgs, text string) {
	msg := tgbotapi.NewMessage(
		args.Update.Message.Chat.ID,
		text,
	)
	if args.NextState != "" && args.User != nil {
		if err := users.UpdateUserState(args.NextState, args.User, args.Ctx, args.UsersCollection); err != nil {
			msg.Text = "Ошибка: " + err.Error()
		}
		msg.ReplyMarkup = states.GetByUserState(args.NextState).GetKeyboard(args.User)
	} else if args.User != nil {
		msg.ReplyMarkup = args.State.GetKeyboard(args.User)
	}
	if _, err := args.Bot.Send(msg); err != nil {
		log.Panic().Err(err).Stack().Msg("Send message panic")
	}
}

func SendAnimation(bot *tgbotapi.BotAPI, chatID int64, fileId string) {
	msg := tgbotapi.NewAnimation(
		chatID,
		tgbotapi.FileID(fileId),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Panic().Err(err).Stack().Msg("Send animantion panic")
	}
}

func SendSticker(bot *tgbotapi.BotAPI, chatID int64, fileId string) {
	msg := tgbotapi.NewSticker(
		chatID,
		tgbotapi.FileID(fileId),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Panic().Err(err).Stack().Msg("Send sticker panic")
	}
}

func SendPhoto(bot *tgbotapi.BotAPI, chatID int64, fileId string) {
	SendPhotoWithCaption(bot, chatID, fileId, "")
}

func SendPhotoWithCaption(bot *tgbotapi.BotAPI, chatID int64, fileId string, caption string) {
	var media []interface{}
	photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileId))
	photo.Caption = caption
	media = append(media, photo)

	mediaGroup := tgbotapi.NewMediaGroup(
		chatID,
		media,
	)
	if _, err := bot.SendMediaGroup(mediaGroup); err != nil {
		log.Panic().Err(err).Stack().Msg("Send photo panic")
	}
}

func CommandStart(args CommandArgs) {
	Send(args, fmt.Sprintf(
		"Привет, %s. Добро пожаловать в Тайного Санту. Чтобы принять участие - зарегистрируйся.",
		args.User.GetName(),
	))
	SendPhoto(args.Bot, args.Update.Message.Chat.ID, ImageStartSantaFileID)
}

func CommandRegistration(args CommandArgs) {
	text := "Ну и как тебя величать?"
	args.NextState = args.State.NextState
	Send(args, text)
}

func CommandMyInfo(args CommandArgs) {
	text := fmt.Sprintf(
		"Тебя зовут: %s\nТвое пожелание: %s",
		args.User.RealName,
		args.User.Wish,
	)
	Send(args, text)
}

func CommandGiftedInfo(args CommandArgs) {
	text := fmt.Sprintf(
		"Ты Тайный Санта для %s\nПожелание: %s",
		args.User.GiftedName,
		args.User.GiftedWish,
	)
	Send(args, text)
}

func CommandChangeName(args CommandArgs) {
	text := "Ну и как тебя величать?"
	args.NextState = users.UserStateChangeName
	Send(args, text)
}

func CommandChangeWish(args CommandArgs) {
	text := "Че те подарить?"
	args.NextState = users.UserStateChangeWish
	Send(args, text)
}

func CommandBack(args CommandArgs) {
	text := "А-аа-ааа-тменяем"
	args.NextState = users.UserStateWait
	Send(args, text)
}

func ErrorText(args CommandArgs) {
	var text string
	if args.State.IsTextAllowed {
		if args.User.State == users.UserStateRegistrationName || args.User.State == users.UserStateChangeName {
			text = "Ты кто, бля? Тебе сказали имя ввести!"
		} else if args.User.State == users.UserStateRegistrationWish || args.User.State == users.UserStateChangeWish {
			text = "Так чего ты хочешь? Определись уже наконец"
		} else {
			text = "Ты че делаешь? Тебе сказали текст ввести!"
		}
	} else {
		text = "Чет херня какая-то. Я меню для кого делал?"
	}
	Send(args, text)
	SendAnimation(args.Bot, args.Update.Message.Chat.ID, AnimationBadSantaWhatFileID)
}

func RegistrationName(args CommandArgs) {
	var text string
	if err := users.UpdateUserRealName(args.Update.Message.Text, args.User, args.Ctx, args.UsersCollection); err != nil {
		text = "Ошибка: " + err.Error()
	} else {
		text = "Что тебе подарить?"
		args.NextState = args.State.NextState
	}
	Send(args, text)
}

func RegistrationWish(args CommandArgs) {
	var text string
	if err := users.UpdateUserWish(args.Update.Message.Text, args.User, args.Ctx, args.UsersCollection); err != nil {
		text = "Ошибка: " + err.Error()
	} else {
		text = "Ты в списке. Жди жеребьевки)"
		args.NextState = args.State.NextState
	}
	Send(args, text)

	SendPhoto(args.Bot, args.Update.Message.Chat.ID, ImageCoolSantaFileID)
}

func ChangeName(args CommandArgs) {
	var text string
	if err := users.UpdateUserRealName(args.Update.Message.Text, args.User, args.Ctx, args.UsersCollection); err != nil {
		text = "Ошибка: " + err.Error()
	} else {
		text = fmt.Sprintf("%s, я тебя запомнил!", args.Update.Message.Text)
		args.NextState = args.State.NextState
	}
	Send(args, text)
	SendAnimation(args.Bot, args.Update.Message.Chat.ID, AnimationChildLooksAtYouFileID)
}

func ChangeWish(args CommandArgs) {
	var text string
	if err := users.UpdateUserWish(args.Update.Message.Text, args.User, args.Ctx, args.UsersCollection); err != nil {
		text = "Ошибка: " + err.Error()
	} else {
		text = "Теперь понятно, чего ты хочешь! Сразу бы так."
		args.NextState = args.State.NextState
	}
	Send(args, text)
	SendAnimation(args.Bot, args.Update.Message.Chat.ID, AnimationChineesePresentFileID)
}

func CommandDrawLots(args CommandArgs) error {
	usersWaiting, err := users.GetUsersByState(users.UserStateWait, args.Ctx, args.UsersCollection)
	if err != nil {
		return err
	}
	log.Printf("Found %d users to send message.", len(usersWaiting))

	completeState := states[users.UserStateComplete]

	if secredSantas, err := getSecretSantas(usersWaiting); err != nil {
		text := "Ошибка: " + err.Error()
		Send(args, text)
	} else {
		for _, santa := range secredSantas {
			presentText := fmt.Sprintf("Хоу-хоу-хоу, %s, твое время пришло!\nТебе предстоит стать Тайным Сантой для %s.\n\nПожелание к подарку: %s.\n\nУдачи, Гильдия Тайных Сант рассчитывает на тебя!",
				santa.Santa.GetName(),
				santa.Gifted.GetNameWithUsername(),
				santa.Gifted.Wish,
			)
			msg := tgbotapi.NewMessage(
				santa.Santa.ChatId,
				presentText,
			)
			msg.ReplyMarkup = completeState.GetKeyboard(santa.Santa)
			if _, err := args.Bot.Send(msg); err != nil {
				log.Panic().Err(err).Stack().Msg("Send message panic")
			}
			SendAnimation(args.Bot, santa.Santa.ChatId, AnimationHolidaySpiritActivatedFileID)
			log.Printf("Message to ID: %d, Usename: %s (%s %s | %s) was sent",
				santa.Santa.TelegramId,
				santa.Santa.Username,
				santa.Santa.FirstName,
				santa.Santa.LastName,
				santa.Santa.RealName,
			)
			if err := users.SetGiftedToUser(santa.Santa, santa.Gifted, args.Ctx, args.UsersCollection); err != nil {
				log.Printf("Update user (%v) error %s", santa.Santa, err.Error())
			}
		}
	}
	return nil
}

func CommandTestDrawLots(args CommandArgs) error {
	usersWaiting, err := users.GetUsersByState(users.UserStateWait, args.Ctx, args.UsersCollection)
	if err != nil {
		return err
	}
	log.Printf("Found %d users to send message.", len(usersWaiting))

	text := ""
	if secredSantas, err := getSecretSantas(usersWaiting); err != nil {
		text += "Ошибка: " + err.Error()
	} else {
		for _, santa := range secredSantas {
			text += santa.Santa.GetNameWithUsername() + " дарит " + santa.Gifted.GetNameWithUsername() + "\n\n"
		}
	}
	Send(args, text)
	return nil
}
