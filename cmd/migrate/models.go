package main

import (
	"time"
)

// MongoDB Models - representing the structure of MongoDB documents

type MongoUser struct {
	ID       int64  `bson:"_id"`
	Language string `bson:"language"`
	Name     string `bson:"name"`
	Username string `bson:"username"`
}

type MongoChat struct {
	ID         int64   `bson:"_id"`
	ChatName   string  `bson:"chat_name"`
	IsInactive bool    `bson:"is_inactive"`
	Language   string  `bson:"language"`
	Users      []int64 `bson:"users"`
}

type MongoAdmin struct {
	ID        int64 `bson:"_id"`
	AnonAdmin bool  `bson:"anon_admin"`
}

type MongoNotesSettings struct {
	ID           int64 `bson:"_id"`
	PrivateNotes bool  `bson:"private_notes"`
}

type MongoNote struct {
	ID          string `bson:"_id"`
	NoteName    string `bson:"note_name"`
	ChatID      int64  `bson:"chat_id"`
	MsgType     int    `bson:"msgtype"`
	NoteContent string `bson:"note_content"`
}

type MongoFilter struct {
	ID          string `bson:"_id"`
	ChatID      int64  `bson:"chat_id"`
	Keyword     string `bson:"keyword"`
	FilterReply string `bson:"filter_reply"`
	MsgType     int    `bson:"msgtype"`
}

type MongoGreeting struct {
	ID              int64                  `bson:"_id"`
	AutoApprove     bool                   `bson:"auto_approve"`
	CleanService    bool                   `bson:"clean_service_settings"`
	GoodbyeSettings map[string]interface{} `bson:"goodbye_settings"`
	WelcomeSettings map[string]interface{} `bson:"welcome_settings"`
}

type MongoLock struct {
	ID           int64                  `bson:"_id"`
	Permissions  map[string]interface{} `bson:"permissions"`
	Restrictions map[string]interface{} `bson:"restrictions"`
}

type MongoPin struct {
	ID             int64 `bson:"_id"`
	AntiChannelPin bool  `bson:"antichannelpin"`
	CleanLinked    bool  `bson:"cleanlinked"`
}

type MongoRule struct {
	ID        int64  `bson:"_id"`
	PrivRules bool   `bson:"privrules"`
	Rules     string `bson:"rules"`
}

type MongoWarnsSetting struct {
	ID        int64  `bson:"_id"`
	WarnLimit int    `bson:"warn_limit"`
	WarnMode  string `bson:"warn_mode"`
}

type MongoWarnsUser struct {
	ID       string   `bson:"_id"`
	ChatID   int64    `bson:"chat_id"`
	UserID   int64    `bson:"user_id"`
	Warns    []string `bson:"warns"`
	NumWarns int      `bson:"num_warns"`
}

type MongoAntifloodSetting struct {
	ID     int64  `bson:"_id"`
	DelMsg bool   `bson:"del_msg"`
	Limit  int64  `bson:"limit"`
	Mode   string `bson:"mode"`
}

type MongoBlacklist struct {
	ID       int64    `bson:"_id"`
	Action   string   `bson:"action"`
	Triggers []string `bson:"triggers"`
	Reason   string   `bson:"reason"`
}

type MongoChannel struct {
	ID          int64  `bson:"_id"`
	ChannelName string `bson:"channel_name"`
	Username    string `bson:"username"`
}

type MongoConnection struct {
	ID        int64 `bson:"_id"`
	Connected bool  `bson:"connected"`
	ChatID    int64 `bson:"chat_id"`
}

type MongoConnectionSetting struct {
	ID         int64 `bson:"_id"`
	CanConnect bool  `bson:"can_connect"`
}

type MongoDisable struct {
	ID           int64    `bson:"_id"`
	Commands     []string `bson:"commands"`
	ShouldDelete bool     `bson:"should_delete"`
}

type MongoReportUserSetting struct {
	ID     int64 `bson:"_id"`
	Status bool  `bson:"status"`
}

type MongoReportChatSetting struct {
	ID     int64 `bson:"_id"`
	Status bool  `bson:"status"`
}

// PostgreSQL Models - representing the structure of PostgreSQL tables

type PgUser struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	UserID    int64  `gorm:"uniqueIndex;not null"`
	Username  string `gorm:"index"`
	Name      string
	Language  string `gorm:"default:'en'"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (PgUser) TableName() string {
	return "users"
}

type PgChat struct {
	ID         int64 `gorm:"primaryKey;autoIncrement"`
	ChatID     int64 `gorm:"uniqueIndex;not null"`
	ChatName   string
	Language   string
	Users      []byte `gorm:"type:jsonb"`
	IsInactive bool   `gorm:"default:false"`
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

func (PgChat) TableName() string {
	return "chats"
}

type PgChatUser struct {
	ChatID int64 `gorm:"primaryKey"`
	UserID int64 `gorm:"primaryKey"`
}

func (PgChatUser) TableName() string {
	return "chat_users"
}

type PgAdmin struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	ChatID    int64 `gorm:"uniqueIndex;not null"`
	AnonAdmin bool  `gorm:"default:false"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (PgAdmin) TableName() string {
	return "admin"
}

type PgNotesSettings struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	ChatID    int64 `gorm:"uniqueIndex;not null"`
	Private   bool  `gorm:"default:false"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (PgNotesSettings) TableName() string {
	return "notes_settings"
}

type PgNote struct {
	ID          int64  `gorm:"primaryKey;autoIncrement"`
	ChatID      int64  `gorm:"not null;index:idx_notes_chat_name"`
	NoteName    string `gorm:"not null;index:idx_notes_chat_name"`
	NoteContent string
	FileID      string
	MsgType     int64
	Buttons     []byte `gorm:"type:jsonb"`
	AdminOnly   bool   `gorm:"default:false"`
	PrivateOnly bool   `gorm:"default:false"`
	GroupOnly   bool   `gorm:"default:false"`
	WebPreview  bool   `gorm:"default:true"`
	IsProtected bool   `gorm:"default:false"`
	NoNotif     bool   `gorm:"default:false"`
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

func (PgNote) TableName() string {
	return "notes"
}

type PgFilter struct {
	ID            int64  `gorm:"primaryKey;autoIncrement"`
	ChatID        int64  `gorm:"not null;index:idx_filters_chat_keyword"`
	Keyword       string `gorm:"not null;index:idx_filters_chat_keyword"`
	FilterReply   string
	Msgtype       int64
	Fileid        string
	Nonotif       bool   `gorm:"default:false"`
	FilterButtons []byte `gorm:"type:jsonb;column:filter_buttons"`
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

func (PgFilter) TableName() string {
	return "filters"
}

type PgGreeting struct {
	ID                   int64 `gorm:"primaryKey;autoIncrement"`
	ChatID               int64 `gorm:"uniqueIndex;not null"`
	CleanServiceSettings bool  `gorm:"default:false"`
	WelcomeCleanOld      bool  `gorm:"default:false"`
	WelcomeLastMsgID     int64
	WelcomeEnabled       bool `gorm:"default:true"`
	WelcomeText          string
	WelcomeFileID        string
	WelcomeType          int64
	WelcomeBtns          []byte `gorm:"type:jsonb"`
	GoodbyeCleanOld      bool   `gorm:"default:false"`
	GoodbyeLastMsgID     int64
	GoodbyeEnabled       bool `gorm:"default:true"`
	GoodbyeText          string
	GoodbyeFileID        string
	GoodbyeType          int64
	GoodbyeBtns          []byte `gorm:"type:jsonb"`
	AutoApprove          bool   `gorm:"default:false"`
	CreatedAt            *time.Time
	UpdatedAt            *time.Time
}

func (PgGreeting) TableName() string {
	return "greetings"
}

type PgLock struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	ChatID    int64  `gorm:"not null;index:idx_lock_chat_type"`
	LockType  string `gorm:"not null;index:idx_lock_chat_type"`
	Locked    bool   `gorm:"default:false"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (PgLock) TableName() string {
	return "locks"
}

type PgPin struct {
	ID             int64 `gorm:"primaryKey;autoIncrement"`
	ChatID         int64 `gorm:"uniqueIndex;not null"`
	MsgID          int64
	CleanLinked    bool `gorm:"default:false"`
	AntiChannelPin bool `gorm:"default:false"`
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
}

func (PgPin) TableName() string {
	return "pins"
}

type PgRule struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	ChatID    int64 `gorm:"uniqueIndex;not null"`
	Rules     string
	RulesBtn  string
	Private   bool `gorm:"default:false"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (PgRule) TableName() string {
	return "rules"
}

type PgWarnsSetting struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	ChatID    int64 `gorm:"uniqueIndex;not null"`
	WarnLimit int64 `gorm:"default:3"`
	WarnMode  string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (PgWarnsSetting) TableName() string {
	return "warns_settings"
}

type PgWarnsUser struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	UserID    int64  `gorm:"not null;index:idx_warns_user_chat"`
	ChatID    int64  `gorm:"not null;index:idx_warns_user_chat"`
	NumWarns  int64  `gorm:"default:0"`
	Warns     []byte `gorm:"type:jsonb"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (PgWarnsUser) TableName() string {
	return "warns_users"
}

type PgAntifloodSetting struct {
	ID                     int64  `gorm:"primaryKey;autoIncrement"`
	ChatID                 int64  `gorm:"uniqueIndex;not null"`
	Limit                  int64  `gorm:"default:5;column:flood_limit"`
	Action                 string `gorm:"default:'mute';column:flood_action"`
	Mode                   string `gorm:"default:'mute'"`
	DeleteAntifloodMessage bool   `gorm:"default:false"`
	CreatedAt              *time.Time
	UpdatedAt              *time.Time
}

func (PgAntifloodSetting) TableName() string {
	return "antiflood_settings"
}

type PgBlacklist struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	ChatID    int64  `gorm:"not null;index:idx_blacklist_chat_word"`
	Word      string `gorm:"not null;index:idx_blacklist_chat_word"`
	Action    string `gorm:"default:'warn'"`
	Reason    string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (PgBlacklist) TableName() string {
	return "blacklists"
}

type PgChannel struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	ChatID    int64 `gorm:"uniqueIndex;not null"`
	ChannelID int64
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (PgChannel) TableName() string {
	return "channels"
}

type PgConnection struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	UserID    int64 `gorm:"not null;index:idx_connection_user_chat"`
	ChatID    int64 `gorm:"not null;index:idx_connection_user_chat"`
	Connected bool  `gorm:"default:false"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (PgConnection) TableName() string {
	return "connection"
}

type PgConnectionSetting struct {
	ID           int64 `gorm:"primaryKey;autoIncrement"`
	ChatID       int64 `gorm:"uniqueIndex;not null"`
	Enabled      bool  `gorm:"default:true"`
	AllowConnect bool  `gorm:"default:true"`
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

func (PgConnectionSetting) TableName() string {
	return "connection_settings"
}

type PgDisable struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	ChatID    int64  `gorm:"not null;index:idx_disable_chat_command"`
	Command   string `gorm:"not null;index:idx_disable_chat_command"`
	Disabled  bool   `gorm:"default:true"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (PgDisable) TableName() string {
	return "disable"
}

type PgReportUserSetting struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	UserID    int64 `gorm:"uniqueIndex;not null"`
	Enabled   bool  `gorm:"default:true"`
	Status    bool  `gorm:"default:true"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (PgReportUserSetting) TableName() string {
	return "report_user_settings"
}

type PgReportChatSetting struct {
	ID          int64  `gorm:"primaryKey;autoIncrement"`
	ChatID      int64  `gorm:"uniqueIndex;not null"`
	Enabled     bool   `gorm:"default:true"`
	Status      bool   `gorm:"default:true"`
	BlockedList []byte `gorm:"type:jsonb"`
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

func (PgReportChatSetting) TableName() string {
	return "report_chat_settings"
}
