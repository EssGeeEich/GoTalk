package nc

import "time"

type NotificationSettings struct {
	ShowUserNotifications    bool // Whether to show notifications for regular 1-on-1 chats
	ShowGroupNotifications   bool // Whether to show notifications for group chats or circles
	ShowBotNotifications     bool // Whether to show notifications for bot chats
	ShowGuestNotifications   bool // Whether to show notifications for guest chats
	ShowBridgedNotifications bool // Whether to show notifications for bridged chats
	ShowMutedNotifications   bool // Force notifications to be shown even if the conversation was muted
	PlayAudio                bool // Plays a notification sound
}

type NotificationSender func(instance string, title string, message string, url string, playAudio bool) error
type NotificationSettingsGetter func() NotificationSettings

type conversationLocalStorage struct {
	lastNotificationTimestamp time.Time
	lastReadMessageId         int64
}
type Monitor struct {
	ncInstance         *Instance
	notificationSender NotificationSender
	settingsGetter     NotificationSettingsGetter
	conversationData   map[int64]conversationLocalStorage
}

func NewMonitor(instance *Instance) *Monitor {
	return &Monitor{
		ncInstance:         instance,
		notificationSender: nil,
		settingsGetter:     nil,
		conversationData:   make(map[int64]conversationLocalStorage),
	}
}

func (m *Monitor) SetNotificationSender(sender NotificationSender) {
	m.notificationSender = sender
}

func (m *Monitor) SetNotificationSettingsGetter(getter NotificationSettingsGetter) {
	m.settingsGetter = getter
}

func (m *Monitor) getNotificationSettings() NotificationSettings {
	if m.settingsGetter != nil {
		return m.settingsGetter()
	}

	return NotificationSettings{
		ShowUserNotifications:    true,
		ShowGroupNotifications:   true,
		ShowBotNotifications:     true,
		ShowGuestNotifications:   true,
		ShowBridgedNotifications: true,
		ShowMutedNotifications:   false,
		PlayAudio:                true,
	}
}

func (m *Monitor) sendMessageNotification(title string, message string, url string, playAudio bool) error {
	if m.notificationSender != nil {
		return m.notificationSender(m.ncInstance.instanceName, title, message, url, playAudio)
	}
	return nil
}

func (m *Monitor) ProcessMessages() (APIResponse, error) {
	conversations, resp, err := m.ncInstance.GetUserConversations()
	if resp != APISuccess || err != nil {
		return resp, err
	}

	activeSettings := m.getNotificationSettings()

	for _, conv := range *conversations {
		var convLocal conversationLocalStorage
		var ok bool
		if convLocal, ok = m.conversationData[conv.Id]; !ok {
			convLocal = conversationLocalStorage{
				lastNotificationTimestamp: time.Unix(0, 0),
				lastReadMessageId:         0,
			}
		}

		if conv.UnreadMessages > 0 {
			if conv.NotificationLevel == 3 && !activeSettings.ShowMutedNotifications {
				continue
			} else if (conv.ActorType == "bots" || conv.LastMessage.ActorType == "bots") && !activeSettings.ShowBotNotifications {
				continue
			} else if (conv.ActorType == "bridged" || conv.LastMessage.ActorType == "bridged") && !activeSettings.ShowBridgedNotifications {
				continue
			} else if (conv.ActorType == "guests" || conv.LastMessage.ActorType == "guests") && !activeSettings.ShowGuestNotifications {
				continue
			} else if (conv.Type == 1 || conv.Type == 5) && !activeSettings.ShowUserNotifications {
				continue
			} else if (conv.Type == 2 || conv.Type == 3) && !activeSettings.ShowGroupNotifications {
				continue
			}

			if time.Since(convLocal.lastNotificationTimestamp).Minutes() > 5 || conv.LastReadMessage != convLocal.lastReadMessageId {
				convLocal.lastNotificationTimestamp = time.Now()
				convLocal.lastReadMessageId = conv.LastReadMessage
				m.conversationData[conv.Id] = convLocal
				defer m.sendMessageNotification(conv.DisplayName, conv.LastMessage.Message, m.ncInstance.GetBaseURL()+"/call/"+conv.LastMessage.Token, activeSettings.PlayAudio)
			}
		}
	}

	return APISuccess, nil
}
