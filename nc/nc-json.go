package nc

import (
	"regexp"
)

type nextcloudLoginFlow struct {
	Poll struct {
		Token    string `json:"token"`
		Endpoint string `json:"endpoint"`
	} `json:"poll"`

	Login string `json:"login"`
}

type nextcloudAuthResult struct {
	Server      string `json:"server"`
	LoginName   string `json:"loginName"`
	AppPassword string `json:"appPassword"`
}

type NextcloudOCSBaseResult[T any] struct {
	OCS struct {
		Meta struct {
			Status     string `json:"status"`
			StatusCode int    `json:"statuscode"`
			Message    string `json:"message"`
		} `json:"meta"`
		Data T `json:"data"`
	} `json:"ocs"`
}

type NextcloudSpreedMessageData struct {
	Id                  int64       `json:"id"`
	Token               string      `json:"token"`
	ActorType           string      `json:"actorType"`
	ActorId             string      `json:"actorId"`
	ActorDisplayName    string      `json:"actorDisplayName"`
	Timestamp           int64       `json:"timestamp"`
	SystemMessage       string      `json:"systemMessage"`
	MessageType         string      `json:"messageType"`
	IsReplyable         bool        `json:"isReplyable"`
	ReferenceId         string      `json:"referenceId"`
	Message             string      `json:"message"`
	MessageParameters   interface{} `json:"messageParameters,omitempty"`
	ExpirationTimestamp int64       `json:"expirationTimestamp"`
	Parent              interface{} `json:"parent,omitempty"`
	Reactions           interface{} `json:"reactions,omitempty"`
	ReactionsSelf       []string    `json:"reactionsSelf,omitempty"`
}

func (msg *NextcloudSpreedMessageData) format() string {
	parameters, ok := msg.MessageParameters.(map[string]interface{})
	if ok && parameters != nil {
		regexReplace := regexp.MustCompile(`{[^{}]*}`)
		return string(regexReplace.ReplaceAllFunc(
			[]byte(msg.Message),
			func(source []byte) []byte {
				// The following line relies on { and } being only 1 byte long.
				// If these characters are replaced with longer characters,
				// the [1 : len(x) - 1] should be applied to the string, not the []byte.
				interfaceParameter, ok := parameters[string(source[1:len(source)-1])]
				if !ok {
					return source
				}
				parameter, ok := interfaceParameter.(map[string]interface{})
				if !ok {
					return source
				}

				interfaceName, ok := parameter["name"]
				if !ok {
					return source
				}
				parameterName, ok := interfaceName.(string)
				if !ok {
					return source
				}

				return []byte(parameterName)
			},
		))
	}
	return msg.Message
}

type NextcloudSpreedConversationData struct {
	Id                    int64                      `json:"id"`
	Token                 string                     `json:"token"`
	Type                  int                        `json:"type"`
	Name                  string                     `json:"name"`
	DisplayName           string                     `json:"displayName"`
	Description           string                     `json:"description"`
	ParticipantType       int                        `json:"participantType"`
	AttendeeId            int                        `json:"attendeeId"`
	AttendeePin           string                     `json:"attendeePin"`
	ActorType             string                     `json:"actorType"`
	ActorId               string                     `json:"actorId"`
	Permissions           int                        `json:"permissions"`
	AttendeePermissions   int                        `json:"attendeePermissions"`
	CallPermissions       int                        `json:"callPermissions"`
	DefaultPermissions    int                        `json:"defaultPermissions"`
	ParticipantFlags      int                        `json:"ParticipantFlags"`
	ReadOnly              int                        `json:"readOnly"`
	Listable              int                        `json:"listable"`
	MessageExpiration     int64                      `json:"messageExpiration"`
	LastPing              int64                      `json:"lastPing"`
	SessionId             string                     `json:"sessionId"`
	HasPassword           bool                       `json:"hasPassword"`
	HasCall               bool                       `json:"hasCall"`
	CallFlag              int                        `json:"callFlag"`
	CanStartCall          bool                       `json:"canStartCall"`
	CanDeleteConversation bool                       `json:"canDeleteConversation"`
	CanLeaveConversation  bool                       `json:"canLeaveConversation"`
	LastActivity          int64                      `json:"lastActivity"`
	IsFavorite            bool                       `json:"isFavorite"`
	NotificationLevel     int                        `json:"notificationLevel"`
	LobbyState            int                        `json:"lobbyState"`
	LobbyTimer            int64                      `json:"lobbyTimer"`
	SipEnabled            int                        `json:"sipEnabled"`
	CanEnableSIP          interface{}                `json:"canEnableSIP"`
	UnreadMessages        int64                      `json:"unreadMessages"`
	UnreadMention         bool                       `json:"unreadMention"`
	UnreadMentionDirect   bool                       `json:"unreadMentionDirect"`
	LastReadMessage       int64                      `json:"lastReadMessage"`
	LastCommonReadMessage int64                      `json:"lastCommonReadMessage"`
	LastMessage           NextcloudSpreedMessageData `json:"lastMessage"`
	ObjectType            string                     `json:"objectType"`
	ObjectId              string                     `json:"objectId"`
	BreakoutRoomMode      interface{}                `json:"breakoutRoomMode"`
	BreakoutRoomStatus    interface{}                `json:"breakoutRoomStatus"`
	Status                string                     `json:"status"`
	StatusIcon            string                     `json:"statusIcon"`
	StatusMessage         string                     `json:"statusMessage"`
	AvatarVersion         string                     `json:"avatarVersion"`
	CallStartTime         int64                      `json:"callStartTime"`
	CallRecording         int                        `json:"callRecording"`
}
