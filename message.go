package fcm

import (
	"errors"
	"strings"
)

var (
	// ErrInvalidMessage occurs if push notitication message is nil.
	ErrInvalidMessage = errors.New("message is invalid")

	// ErrInvalidTarget occurs if message topic is empty.
	ErrInvalidTarget = errors.New("topic is invalid or registration ids are not set")

	// ErrToManyRegIDs occurs when registration ids more then 1000.
	ErrToManyRegIDs = errors.New("too many registrations ids")

	// ErrInvalidTimeToLive occurs if TimeToLive more then 2419200.
	ErrInvalidTimeToLive = errors.New("messages time-to-live is invalid")
)

// Notification specifies the predefined, user-visible key-value pairs of the
// notification payload.
type Notification struct {
	Title        string `json:"title,omitempty"`
	Body         string `json:"body,omitempty"`
	ChannelID    string `json:"android_channel_id,omitempty"`
	Icon         string `json:"icon,omitempty"`
	Image        string `json:"image,omitempty"`
	Sound        string `json:"sound,omitempty"`
	Badge        string `json:"badge,omitempty"`
	Tag          string `json:"tag,omitempty"`
	Color        string `json:"color,omitempty"`
	ClickAction  string `json:"click_action,omitempty"`
	BodyLocKey   string `json:"body_loc_key,omitempty"`
	BodyLocArgs  string `json:"body_loc_args,omitempty"`
	TitleLocKey  string `json:"title_loc_key,omitempty"`
	TitleLocArgs string `json:"title_loc_args,omitempty"`
}

// Message represents list of targets, options, and payload for HTTP JSON
// messages.
type Message struct {
	Topic                 string                 `json:"topic,omitempty"` // Changed from "to" to "topic"
	Token                 string                 `json:"token,omitempty"`
	Condition             string                 `json:"condition,omitempty"`
	CollapseKey           string                 `json:"collapse_key,omitempty"`
	Priority              string                 `json:"priority,omitempty"`
	ContentAvailable      bool                   `json:"content_available,omitempty"`
	MutableContent        bool                   `json:"mutable_content,omitempty"`
	TimeToLive            *uint                  `json:"time_to_live,omitempty"`
	DryRun                bool                   `json:"dry_run,omitempty"`
	RestrictedPackageName string                 `json:"restricted_package_name,omitempty"`
	Notification          *Notification          `json:"notification,omitempty"`
	Data                  map[string]interface{} `json:"data,omitempty"`
	Apns                  map[string]interface{} `json:"apns,omitempty"`
	Webpush               map[string]interface{} `json:"webpush,omitempty"`
}

type NewMessage struct {
	Message Message `json:"message"`
}

// Validate returns an error if the message is not well-formed.
func (msg *NewMessage) Validate() error {
	if msg == nil {
		return ErrInvalidMessage
	}

	// validate target identifier: `token`, `topic`, or `condition`
	opCnt := strings.Count(msg.Message.Condition, "&&") + strings.Count(msg.Message.Condition, "||")
	if msg.Message.Token == "" && msg.Message.Topic == "" && (msg.Message.Condition == "" || opCnt > 5) {
		return ErrInvalidTarget
	}

	if msg.Message.TimeToLive != nil && *msg.Message.TimeToLive > uint(2419200) {
		return ErrInvalidTimeToLive
	}
	return nil
}
