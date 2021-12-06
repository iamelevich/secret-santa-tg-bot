package updates

import (
	"fmt"
	"math/rand"
	"secret-santa-go-bot/pkg/users"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
)

func shuffle(array []users.User) []users.User {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	newArray := array
	r1.Shuffle(len(newArray), func(i, j int) { newArray[i], newArray[j] = newArray[j], newArray[i] })
	return newArray
}

type SecretSanta struct {
	Santa  *users.User
	Gifted *users.User
}

func getSecretSantas(usrs []users.User) ([]SecretSanta, error) {
	var santas []SecretSanta
	sortedUsers := make([]users.User, len(usrs))
	_ = copy(sortedUsers, usrs)
	sort.Sort(users.ByExcludeIds(sortedUsers))

	log.Debug().
		Strs("sorted_users", users.UsersAsStrings(sortedUsers)).
		Msg("Sort by exclude ID")

	giftedUsers := make([]users.User, len(usrs))
	_ = copy(giftedUsers, sortedUsers)

	for i := range sortedUsers {
		santa := SecretSanta{
			Santa: &sortedUsers[i],
		}
		if gifted, index, err := getGiftedForSanta(santa.Santa, giftedUsers); err != nil {
			return santas, err
		} else {
			santa.Gifted = gifted
			giftedUsers = append(giftedUsers[:index], giftedUsers[index+1:]...)
		}
		santas = append(santas, santa)
	}

	return santas, nil
}

func getGiftedForSanta(santa *users.User, giftedList []users.User) (*users.User, int, error) {
	var filtered []users.User
	var excludedIds = santa.GetExcludedIds()

	for _, gifted := range giftedList {
		excluded := false
		for _, excludeId := range excludedIds {
			if fmt.Sprintf("%d", gifted.TelegramId) == excludeId || gifted.ID == santa.ID {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, gifted)
		}
	}

	log.Debug().
		Str("santa", santa.GetNameWithUsername()).
		Strs("variants", users.UsersAsStrings(filtered)).
		Msg("Try to find gifted")

	if len(filtered) == 0 {
		return nil, 0, fmt.Errorf("Can't find gifted for santa %s", santa.GetNameWithUsername())
	}

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	giftedUser := &filtered[r.Intn(len(filtered))]
	index := 0
	for i, gifted := range giftedList {
		if gifted.ID == giftedUser.ID {
			index = i
			break
		}
	}
	return giftedUser, index, nil
}
