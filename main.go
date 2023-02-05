package main

import (
	"log"
	"time"

	"GoTalk/nc"
	"GoTalk/ncmonitor"
	"GoTalk/settings"

	"github.com/billgraziano/dpapi"
	"gopkg.in/toast.v1"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/systray"
)

type conversationLocalStorage struct {
	lastNotificationTimestamp time.Time
	lastReadMessageId         int64
}

var (
	settingsManager *settings.SettingsManager[UserSettings, OrgSettings]
	user            *UserSettings
	org             *OrgSettings
	convLocalData   map[int64]conversationLocalStorage
)

func sendMessageNotification(title string, message string, url string) error {
	notification := toast.Notification{
		AppID:               "Nextcloud Talk",
		Title:               title,
		Message:             message,
		Audio:               toast.Silent,
		ActivationArguments: url,
		Icon:                "DefaultIcon.png",
	}

	if user.PlayAudio {
		notification.Audio = toast.IM
	}

	if err := notification.Push(); err != nil {
		return err
	}

	return nil
}

func runNextcloudMonitor() error {
	timer := time.NewTicker(5 * time.Second)

	var ncCredentials nc.AuthCredentials

	{
		var decPassword string
		var err error
		if decPassword, err = dpapi.Decrypt(user.AppPassword); err != nil {
			decPassword = ""
		}

		ncCredentials = nc.AuthCredentials{
			LoginName:   user.Username,
			AppPassword: decPassword,
		}
	}

	ncInstance := nc.NewInstance(org.NextcloudInstance)
	ncInstance.SetCredentials(ncCredentials)

	ncMonitor := ncmonitor.NewMonitor(ncInstance)
	ncMonitor.SetNotificationSender(sendMessageNotification)

	for range timer.C {
		needsLogin, err := ncMonitor.NeedsLogin()
		if err != nil {
			return err
		}

		if needsLogin {
			// ...
		} else if err := ncMonitor.ProcessMessages(); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	var err error
	convLocalData = make(map[int64]conversationLocalStorage)
	settingsManager := settings.NewSettingsManager(
		"SGH",
		"GoTalk",

		// Default User Settings
		UserSettings{
			PlayAudio:         true,
			ShowNotifications: true,
		},

		// Default Org Settings
		OrgSettings{
			CanAddInstances:  false,
			MessageCheckTime: 5,
		},
	)

	if user, org, err = settingsManager.Load(); err != nil {
		log.Fatal(err)
		return
	}

	defer func() {
		if err := settingsManager.Save(user, org); err != nil {
			log.Fatal(err)
			return
		}
	}()

	icon, _ := fyne.LoadResourceFromPath("DefaultIcon.ico")
	a := app.New()

	if desk, ok := a.(desktop.App); ok {
		var menu *fyne.Menu
		var showUserNotifications *fyne.MenuItem
		var showGroupNotifications *fyne.MenuItem
		var showBotNotifications *fyne.MenuItem
		var showGuestNotifications *fyne.MenuItem
		var showBridgedNotifications *fyne.MenuItem
		var showMutedNotifications *fyne.MenuItem
		var playAudioSound *fyne.MenuItem

		updateSettings := func() {
			user.ShowUserNotifications = showUserNotifications.Checked
			user.ShowGroupNotifications = showGroupNotifications.Checked
			user.ShowBotNotifications = showBotNotifications.Checked
			user.ShowGuestNotifications = showGuestNotifications.Checked
			user.ShowBridgedNotifications = showBridgedNotifications.Checked
			user.ShowMutedNotifications = showMutedNotifications.Checked
			user.PlayAudio = playAudioSound.Checked
		}

		readbackSettings := func() {
			showUserNotifications.Checked = user.ShowUserNotifications
			showGroupNotifications.Checked = user.ShowGroupNotifications
			showBotNotifications.Checked = user.ShowBotNotifications
			showGuestNotifications.Checked = user.ShowGuestNotifications
			showBridgedNotifications.Checked = user.ShowBridgedNotifications
			showMutedNotifications.Checked = user.ShowMutedNotifications
			playAudioSound.Checked = user.PlayAudio

			if menu != nil {
				menu.Refresh()
			}
		}

		showUserNotifications = fyne.NewMenuItem("Show User Notifications", func() {
			showUserNotifications.Checked = !showUserNotifications.Checked
			updateSettings()
			menu.Refresh()
		})
		showGroupNotifications = fyne.NewMenuItem("Show Group Notifications", func() {
			showGroupNotifications.Checked = !showGroupNotifications.Checked
			updateSettings()
			menu.Refresh()
		})
		showBotNotifications = fyne.NewMenuItem("Show Bot Notifications", func() {
			showBotNotifications.Checked = !showBotNotifications.Checked
			updateSettings()
			menu.Refresh()
		})
		showGuestNotifications = fyne.NewMenuItem("Show Guest Notifications", func() {
			showGuestNotifications.Checked = !showGuestNotifications.Checked
			updateSettings()
			menu.Refresh()
		})
		showBridgedNotifications = fyne.NewMenuItem("Show Bridged Notifications", func() {
			showBridgedNotifications.Checked = !showBridgedNotifications.Checked
			updateSettings()
			menu.Refresh()
		})
		showMutedNotifications = fyne.NewMenuItem("Show Muted Notifications", func() {
			showMutedNotifications.Checked = !showMutedNotifications.Checked
			updateSettings()
			menu.Refresh()
		})
		playAudioSound = fyne.NewMenuItem("Notification Sound", func() {
			playAudioSound.Checked = !playAudioSound.Checked
			updateSettings()
			menu.Refresh()
		})

		readbackSettings()

		menu = fyne.NewMenu(
			"GoTalk",
			showUserNotifications,
			showGroupNotifications,
			showBotNotifications,
			showGuestNotifications,
			showBridgedNotifications,
			fyne.NewMenuItemSeparator(),
			showMutedNotifications,
			playAudioSound,
		)

		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayIcon(icon)
	}

	go func() {
		err := runNextcloudMonitor()
		if err != nil {
			log.Print(err)
		}
		a.Quit()
	}()

	defer systray.Quit()

	a.Run()
}
