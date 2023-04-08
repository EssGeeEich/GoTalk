package nc

import "time"

type NotificationSettings struct {
	ShowUserNotifications    bool // Whether to show notifications for regular 1-on-1 chats
	ShowGroupNotifications   bool // Whether to show notifications for group chats or circles
	ShowBotNotifications     bool // Whether to show notifications for bot chats
	ShowGuestNotifications   bool // Whether to show notifications for guest chats
	ShowBridgedNotifications bool // Whether to show notifications for bridged chats
	ShowMutedNotifications   bool // Force notifications to be shown even if the conversation was muted
	PlayNotificationSounds   bool // Plays a notification sound
}

type NotificationSender func(instance string, title string, message string, url string, playAudio bool) error
type NotificationCountSetter func(instance string, unfilteredCount uint, filteredCount uint) error
type NotificationSettingsGetter func() NotificationSettings

type conversationLocalStorage struct {
	lastNotificationTimestamp time.Time
	lastMessageId             int64
}

type Monitor struct {
	ncInstance              *Instance
	repeatTime              float64
	notificationSender      NotificationSender
	notificationCountSetter NotificationCountSetter
	settingsGetter          NotificationSettingsGetter
	conversationData        map[int64]conversationLocalStorage
}

func NewMonitor(instance *Instance, repeatTime float64) *Monitor {
	return &Monitor{
		ncInstance:         instance,
		repeatTime:         repeatTime,
		notificationSender: nil,
		settingsGetter:     nil,
		conversationData:   make(map[int64]conversationLocalStorage),
	}
}

func (m *Monitor) SetNotificationSender(sender NotificationSender) {
	m.notificationSender = sender
}

func (m *Monitor) SetNotificationCountSetter(setter NotificationCountSetter) {
	m.notificationCountSetter = setter
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
		PlayNotificationSounds:   true,
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

	var filteredCount uint = 0
	var unfilteredCount uint = 0
	for _, conv := range *conversations {
		var convLocal conversationLocalStorage
		var ok bool
		if convLocal, ok = m.conversationData[conv.Id]; !ok {
			convLocal = conversationLocalStorage{
				lastNotificationTimestamp: time.Unix(0, 0),
				lastMessageId:             0,
			}
		}

		if conv.UnreadMessages > 0 && conv.LastMessage.ActorId != conv.ActorId {
			unfilteredCount += 1
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
			filteredCount += 1

			minsSinceLastNotification := time.Since(convLocal.lastNotificationTimestamp).Minutes()
			if (minsSinceLastNotification >= 0.5 && minsSinceLastNotification >= m.repeatTime) || conv.LastMessage.Id != convLocal.lastMessageId {
				textPreview := conv.LastMessage.format()
				convLocal.lastNotificationTimestamp = time.Now()
				convLocal.lastMessageId = conv.LastMessage.Id
				m.conversationData[conv.Id] = convLocal
				defer m.sendMessageNotification(conv.DisplayName, textPreview, m.ncInstance.GetBaseURL()+"/call/"+conv.LastMessage.Token, activeSettings.PlayNotificationSounds)
			}
		}
	}
	if m.notificationCountSetter != nil {
		m.notificationCountSetter(m.ncInstance.instanceName, unfilteredCount, filteredCount)
	}

	return APISuccess, nil
}
