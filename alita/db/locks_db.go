package db

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Locks represents all lock and restriction settings for a chat.
//
// Fields:
//   - ChatId: Unique identifier for the chat.
//   - Permissions: Permissions for various message types and actions.
//   - Restrictions: Additional restrictions for the chat.
type Locks struct {
	ChatId       int64         `bson:"_id,omitempty" json:"_id,omitempty"`
	Permissions  *Permissions  `bson:"permissions,omitempty" json:"permissions,omitempty"`
	Restrictions *Restrictions `bson:"restrictions,omitempty" json:"restrictions,omitempty"`
}

// Permissions defines which message types and actions are locked in a chat.
type Permissions struct {
	Sticker       bool `bson:"sticker,omitempty" json:"sticker,omitempty"`
	Audio         bool `bson:"audio,omitempty" json:"audio,omitempty"`
	Voice         bool `bson:"voice,omitempty" json:"voice,omitempty"`
	Video         bool `bson:"video,omitempty" json:"video,omitempty"`
	Document      bool `bson:"document,omitempty" json:"document,omitempty"`
	VideoNote     bool `bson:"video_note,omitempty" json:"video_note,omitempty"`
	Contact       bool `bson:"contact,omitempty" json:"contact,omitempty"`
	Photo         bool `bson:"photo,omitempty" json:"photo,omitempty"`
	Gif           bool `bson:"gif,omitempty" json:"gif,omitempty"`
	Url           bool `bson:"url,omitempty" json:"url,omitempty"`
	Bot           bool `bson:"bot,omitempty" json:"bot,omitempty"`
	Forward       bool `bson:"forward,omitempty" json:"forward,omitempty"`
	Game          bool `bson:"game,omitempty" json:"game,omitempty"`
	Location      bool `bson:"location,omitempty" json:"location,omitempty"`
	Arab          bool `bson:"arab_chars,omitempty" json:"arab_chars,omitempty"`
	SendAsChannel bool `bson:"send_as_channel,omitempty" json:"send_as_channel,omitempty"`
}

// Restrictions defines additional restrictions for a chat.
type Restrictions struct {
	Messages        bool `bson:"messages,omitempty" json:"messages,omitempty"`
	ChannelComments bool `bson:"channel_comments,omitempty" json:"channel_comments,omitempty"`
	Media           bool `bson:"media,omitempty" json:"media,omitempty"`
	Other           bool `bson:"other,omitempty" json:"other,omitempty"`
	Previews        bool `bson:"previews,omitempty" json:"previews,omitempty"`
	All             bool `bson:"all,omitempty" json:"all,omitempty"`
}

// checkChatLocks fetches lock settings for a chat from the database.
// If no document exists, it creates one with default values (all locks disabled).
// Returns a pointer to the Locks struct with either existing or default values.
func checkChatLocks(chatID int64) (lockrc *Locks) {
	defaultLockrc := &Locks{ChatId: chatID, Permissions: &Permissions{}, Restrictions: &Restrictions{}}
	errS := findOne(lockColl, bson.M{"_id": chatID}).Decode(&lockrc)
	if errS == mongo.ErrNoDocuments {
		lockrc = defaultLockrc
		err := updateOne(lockColl, bson.M{"_id": chatID}, lockrc)
		if err != nil {
			log.Errorf("[Database] checkChatLocks: %v", err)
		}
	} else if errS != nil {
		log.Errorf("[Database][checkChatLocks]: %v", errS)
		lockrc = defaultLockrc
	}
	return lockrc
}

// GetChatLocks retrieves the lock settings for a given chat ID.
// If no settings exist, it initializes them with default values (all locks disabled).
// This is the main function for accessing lock settings.
func GetChatLocks(chatID int64) *Locks {
	return checkChatLocks(chatID)
}

// MapLockType returns a map of lock and restriction types to their enabled/disabled state for a chat.
// Converts the structured lock settings into a flat map for easier querying.
// Used by other functions to check specific lock states.
func MapLockType(locks Locks) map[string]bool {
	perms := locks.Permissions
	restr := locks.Restrictions
	m := map[string]bool{
		"sticker":     perms.Sticker,
		"audio":       perms.Audio,
		"voice":       perms.Voice,
		"document":    perms.Document,
		"video":       perms.Video,
		"videonote":   perms.VideoNote,
		"contact":     perms.Contact,
		"photo":       perms.Photo,
		"gif":         perms.Gif,
		"url":         perms.Url,
		"bots":        perms.Bot,
		"forward":     perms.Forward,
		"game":        perms.Game,
		"location":    perms.Location,
		"rtl":         perms.Arab,
		"anonchannel": perms.SendAsChannel,
		"messages":    restr.Messages,
		"comments":    restr.ChannelComments,
		"media":       restr.Media,
		"other":       restr.Other,
		"previews":    restr.Previews,
		"all":         restr.All,
	}
	return m
}

// UpdateLock modifies the value of a specific lock or restriction for a chat and updates it in the database.
// The perm argument specifies which lock/restriction to update (e.g., "sticker", "photo", "url").
// When enabled (val=true), the specified content type becomes restricted in the chat.
func UpdateLock(chatID int64, perm string, val bool) {
	lockrc := checkChatLocks(chatID)

	switch perm {
	case "sticker":
		lockrc.Permissions.Sticker = val
	case "audio":
		lockrc.Permissions.Audio = val
	case "voice":
		lockrc.Permissions.Voice = val
	case "document":
		lockrc.Permissions.Document = val
	case "video":
		lockrc.Permissions.Video = val
	case "videonote":
		lockrc.Permissions.VideoNote = val
	case "contact":
		lockrc.Permissions.Contact = val
	case "photo":
		lockrc.Permissions.Photo = val
	case "gif":
		lockrc.Permissions.Gif = val
	case "url":
		lockrc.Permissions.Url = val
	case "bots":
		lockrc.Permissions.Bot = val
	case "forward":
		lockrc.Permissions.Forward = val
	case "game":
		lockrc.Permissions.Game = val
	case "location":
		lockrc.Permissions.Location = val
	case "rtl":
		lockrc.Permissions.Arab = val
	case "anonchannel":
		lockrc.Permissions.SendAsChannel = val
	case "messages":
		lockrc.Restrictions.Messages = val
	case "comments":
		lockrc.Restrictions.ChannelComments = val
	case "media":
		lockrc.Restrictions.Media = val
	case "other":
		lockrc.Restrictions.Other = val
	case "previews":
		lockrc.Restrictions.Previews = val
	case "all":
		lockrc.Restrictions.All = val
	}

	err := updateOne(lockColl, bson.M{"_id": chatID}, lockrc)
	if err != nil {
		log.Errorf("[Database] UpdateLock: %v", err)
	}
}

// IsPermLocked returns true if the specified permission or restriction is locked for the chat.
// Uses MapLockType internally to check the lock state efficiently.
// Returns false for unknown permission types.
func IsPermLocked(chatID int64, perm string) bool {
	lockrc := checkChatLocks(chatID)
	return MapLockType(*lockrc)[perm]
}
