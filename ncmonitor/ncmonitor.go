package ncmonitor

import (
	"GoTalk/nc"
)

type NotificationSender func(title string, message string, url string) error

type Monitor struct {
	notif NotificationSender
}

func NewMonitor(instance *nc.Instance) *Monitor {

}

func (m *Monitor) SetNotificationSender(sender NotificationSender) {
	m.notif = sender
}

func (m *Monitor) NeedsLogin() (bool, error) {

}

func (m *Monitor) ProcessMessages() error {

}

/*
if !nextcloudInstanceValid {
	nextcloudInstanceValid, err = startNextcloudInstance()
	if err != nil {
		return err
	}

	if err = mySettings.Save("GoTalk.toml"); err != nil {
		return err
	}
} else {
	nextcloudInstanceValid, err = updateNextcloudInstance()
	if err != nil {
		return err
	}
}

func startNextcloudInstance() (bool, error) {
	selCredentials := &nc.AuthCredentials{
		LoginName:   mySettings.Get().Username,
		AppPassword: mySettings.Get().AppPassword,
	}

	result, err := ncInstance.ValidateCredentials(*selCredentials)
	if err != nil {
		return false, err
	}

	if result {
		ncInstance.SetCredentials(*selCredentials)
		return true, nil
	}

	ncLoginFlow := ncInstance.NewLoginFlow()
	if err := ncLoginFlow.Start(); err != nil {
		return false, err
	}

	if selCredentials, err = ncLoginFlow.WaitFlow(); err != nil {
		return false, err
	}

	if result, err = ncInstance.ValidateCredentials(*selCredentials); err != nil {
		return false, err
	}

	if result {
		s := mySettings.Get()
		s.Username = selCredentials.LoginName
		s.AppPassword = selCredentials.AppPassword
		mySettings.Set(s)

		ncInstance.SetCredentials(*selCredentials)
		return true, nil
	}

	return false, errors.New("login flow didn't respond with a valid token")
}

func updateNextcloudInstance() (bool, error) {
	conversations, err := ncInstance.GetUserConversations()
	if err != nil {
		s, _ := ncInstance.ValidateCredentials(ncInstance.GetCredentials())
		return s, err
	}

	activeSettings := mySettings.Get()

	for _, conv := range *conversations {
		var convLocal conversationLocalStorage
		var ok bool
		if convLocal, ok = convLocalData[conv.Id]; !ok {
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
				convLocalData[conv.Id] = convLocal
				defer sendMessageNotification(conv.DisplayName, conv.LastMessage.Message, ncInstance.GetBaseURL()+"/call/"+conv.LastMessage.Token, activeSettings.PlayAudio)
			}
		}
	}

	return true, nil
}
*/
