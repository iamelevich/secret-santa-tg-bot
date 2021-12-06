package updates

import (
	"secret-santa-go-bot/pkg/users"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StateData struct {
	KeyboardButtons [][]string
	AdminButtons    []string
	IsTextAllowed   bool
	NextState       string
}

func (sd *StateData) GetKeyboard(user *users.User) interface{} {
	var keyboard [][]tgbotapi.KeyboardButton

	for _, kbRow := range sd.KeyboardButtons {
		var keyboardRow []tgbotapi.KeyboardButton
		for _, kb := range kbRow {
			keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButton(kb))
		}
		keyboard = append(keyboard, keyboardRow)
	}

	if user.IsAdmin && len(sd.AdminButtons) > 0 {
		var keyboardRow []tgbotapi.KeyboardButton
		for _, kb := range sd.AdminButtons {
			keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButton(kb))
		}
		keyboard = append(keyboard, keyboardRow)
	}

	if len(keyboard) == 0 {
		return tgbotapi.ReplyKeyboardRemove{
			RemoveKeyboard: true,
		}
	}

	return tgbotapi.ReplyKeyboardMarkup{
		ResizeKeyboard:  true,
		Keyboard:        keyboard,
		OneTimeKeyboard: false,
	}
}

func (sd *StateData) HasOption(user *users.User, option string) bool {
	exists := false
	for _, kbr := range sd.KeyboardButtons {
		for _, kb := range kbr {
			if kb == option {
				exists = true
				break
			}
		}
		if exists {
			break
		}
	}
	if user.IsAdmin {
		for _, kb := range sd.AdminButtons {
			if kb == option {
				exists = true
				break
			}
		}
	}
	return exists
}

type StatesData map[string]StateData

func (sd *StatesData) GetByUserState(userState string) *StateData {
	var state StateData
	if _, prs := states[userState]; prs {
		state = states[userState]
	} else {
		state = states["start"]
	}
	return &state
}
