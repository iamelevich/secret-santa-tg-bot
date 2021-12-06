package users

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	ID         primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	TelegramId int64              `bson:"telegram_id"`
	Username   string             `bson:"username"`
	FirstName  string             `bson:"first_name"`
	LastName   string             `bson:"last_name"`
	State      string             `bson:"state"`
	IsAdmin    bool               `bson:"is_admin"`
	ChatId     int64              `bson:"chat_id"`
	ExcludeIds string             `bson:"exclude_ids"`

	RealName   string `bson:"real_name"`
	Wish       string `bson:"wish"`
	GiftedName string `bson:"gifted_name"`
	GiftedWish string `bson:"gifted_wish"`
}

func (u *User) GetName() string {
	name := "Незнакомец"
	if u.RealName != "" {
		name = u.RealName
	} else if u.FirstName != "" && u.LastName != "" {
		name = fmt.Sprintf("%s %s", u.FirstName, u.LastName)
	} else if u.FirstName != "" {
		name = u.FirstName
	} else if u.LastName != "" {
		name = u.LastName
	} else if u.Username != "" {
		name = u.Username
	}
	return name
}

func (u *User) GetNameWithUsername() string {
	if u.Username != "" {
		return fmt.Sprintf("%s (@%s)", u.GetName(), u.Username)
	}
	return u.GetName()
}

func (u *User) GetExcludedIds() []string {
	if u.ExcludeIds == "" {
		return []string{}
	}
	return strings.Split(u.ExcludeIds, ",")
}

type ByExcludeIds []User

func (a ByExcludeIds) Len() int      { return len(a) }
func (a ByExcludeIds) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByExcludeIds) Less(i, j int) bool {
	return len(a[i].GetExcludedIds()) > len(a[j].GetExcludedIds())
}

const (
	UserStateStart            string = "start"
	UserStateRegistrationName string = "registration_name"
	UserStateRegistrationWish string = "registration_wish"
	UserStateWait             string = "wait"
	UserStateChangeName       string = "change_name"
	UserStateChangeWish       string = "change_wish"
	UserStateComplete         string = "complete"
)

func GetUsersCollection(client *mongo.Client) *mongo.Collection {
	return client.Database(os.Getenv("MONGODB_DATABASE")).Collection("users")
}

func CreateUser(
	pCtx context.Context,
	usersCollection *mongo.Collection,
	update *tgbotapi.Update,
) (User, error) {
	var user User
	ctx, cancel := context.WithTimeout(pCtx, 5*time.Second)
	defer cancel()
	user = User{
		ID:         primitive.NewObjectID(),
		TelegramId: update.Message.From.ID,
		Username:   update.Message.From.UserName,
		FirstName:  update.Message.From.FirstName,
		LastName:   update.Message.From.LastName,
		State:      UserStateStart,
		IsAdmin:    false,
		ChatId:     update.Message.Chat.ID,
	}
	bUser, err := bson.Marshal(user)
	if err != nil {
		return user, err
	}
	res, err := usersCollection.InsertOne(ctx, bUser)
	if err != nil {
		return user, err
	}
	user.ID = res.InsertedID.(primitive.ObjectID)
	return user, nil
}

func GetUser(
	pCtx context.Context,
	usersCollection *mongo.Collection,
	update *tgbotapi.Update,
) (User, error) {
	var user User
	ctx, cancel := context.WithTimeout(pCtx, 5*time.Second)
	defer cancel()
	err := usersCollection.FindOne(ctx, bson.D{{Key: "telegram_id", Value: update.Message.From.ID}}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		user, err = CreateUser(pCtx, usersCollection, update)
		if err != nil {
			return user, err
		}
	} else if err != nil {
		return user, err
	}
	return user, nil
}

func UpdateUserState(
	newState string,
	user *User,
	pCtx context.Context,
	usersCollection *mongo.Collection,
) error {
	ctx, cancel := context.WithTimeout(pCtx, 5*time.Second)
	defer cancel()
	_, err := usersCollection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: user.ID}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "state", Value: newState}}}},
	)
	return err
}

func UpdateUserRealName(
	newRealName string,
	user *User,
	pCtx context.Context,
	usersCollection *mongo.Collection,
) error {
	ctx, cancel := context.WithTimeout(pCtx, 5*time.Second)
	defer cancel()
	_, err := usersCollection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: user.ID}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "real_name", Value: newRealName}}}},
	)
	return err
}

func UpdateUserWish(
	newWish string,
	user *User,
	pCtx context.Context,
	usersCollection *mongo.Collection,
) error {
	ctx, cancel := context.WithTimeout(pCtx, 5*time.Second)
	defer cancel()
	_, err := usersCollection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: user.ID}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "wish", Value: newWish}}}},
	)
	return err
}

func GetUsersByState(
	state string,
	pCtx context.Context,
	usersCollection *mongo.Collection,
) ([]User, error) {
	var users []User
	ctx, cancel := context.WithTimeout(pCtx, 10*time.Second)
	defer cancel()
	cursor, err := usersCollection.Find(ctx, bson.D{{Key: "state", Value: state}})
	if err != nil {
		return users, err
	}
	if err = cursor.All(ctx, &users); err != nil {
		return users, err
	}
	return users, nil
}

func SetGiftedToUser(
	user *User,
	gifted *User,
	pCtx context.Context,
	usersCollection *mongo.Collection,
) error {
	ctx, cancel := context.WithTimeout(pCtx, 5*time.Second)
	defer cancel()
	_, err := usersCollection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: user.ID}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "state", Value: UserStateComplete},
			{Key: "gifted_name", Value: gifted.GetNameWithUsername()},
			{Key: "gifted_wish", Value: gifted.Wish},
		}}},
	)
	return err
}

func UsersAsStrings(usrs []User) []string {
	var result []string
	for _, user := range usrs {
		result = append(
			result,
			fmt.Sprintf("%d: %s (%s)", user.TelegramId, user.GetNameWithUsername(), user.ExcludeIds),
		)
	}
	return result
}
