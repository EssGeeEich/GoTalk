package main

import (
	"log"
	"os"
	"sort"
	"sync"

	"GoTalk/nc"
	"GoTalk/settings"

	"gopkg.in/toast.v1"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/systray"
)

var (
	cache *Cache
	user  *UserSettings
	org   *OrgSettings

	settingsManager *settings.SettingsManager[Cache, UserSettings, OrgSettings]
)

func sendMessageNotification(instance string, title string, message string, url string, playAudio bool) error {
	if !user.ShowNotifications {
		return nil
	}

	var err error

	// cacheInstance, cacheInstanceOk := cache.InstanceData[instance]
	orgInstance, orgInstanceOk := org.InstanceData[instance]

	// Determine which icon should be displayed
	var icon string
	if orgInstanceOk && orgInstance.NotificationAppIcon != "" {
		icon = orgInstance.NotificationAppIcon
	} else {
		if icon, err = os.Executable(); err != nil {
			return err
		}

		icon = icon + string(os.PathSeparator) + "DefaultIcon.png"
	}

	notification := toast.Notification{
		AppID:               "Nextcloud Talk",
		Title:               title,
		Message:             message,
		Audio:               toast.IM,
		ActivationArguments: url,
		Icon:                icon,
		Actions: []toast.Action{
			{Type: "protocol", Label: "Open", Arguments: url},
		},
	}

	// Determine whether the user wants audio for this instance
	if !user.PlayNotificationSounds || !playAudio {
		notification.Audio = toast.Silent
	}

	if err = notification.Push(); err != nil {
		return err
	}

	return nil
}

func startNextcloudMonitor(wg *sync.WaitGroup, closeChan chan interface{}) error {
	for instanceName := range org.InstanceData {

		// Run the actual Monitor loop
		wg.Add(1)
		internalMonitor := newMonitorProc(instanceName)
		go internalMonitor.run(wg, closeChan)
	}

	return nil
}

func main() {
	var err error
	settingsManager = settings.NewSettingsManager(
		"SGH",
		"GoTalk",

		Cache{},

		// Default User Settings
		UserSettings{
			PlayNotificationSounds: true,
			ShowNotifications:      true,
		},

		// Default Org Settings
		OrgSettings{
			MessageCheckTime: 5,
		},
	)

	if cache, user, org, err = settingsManager.Load(); err != nil {
		log.Fatal(err)
		return
	}

	if cache.InstanceData == nil {
		cache.InstanceData = make(map[string]InstanceCache)
	}

	if user.InstanceData == nil {
		user.InstanceData = make(map[string]UserInstanceSettings)
	}

	if org.InstanceData == nil {
		org.InstanceData = make(map[string]OrgInstanceSettings)
	}

	for instanceName := range org.InstanceData {
		if _, ok := user.InstanceData[instanceName]; !ok {
			// Sensible default user settings for a new instance
			user.InstanceData[instanceName] = UserInstanceSettings{
				NotificationSettings: nc.NotificationSettings{
					ShowUserNotifications:    true,
					ShowGroupNotifications:   true,
					ShowBotNotifications:     true,
					ShowGuestNotifications:   true,
					ShowBridgedNotifications: true,
					ShowMutedNotifications:   false,
					PlayNotificationSounds:   true,
				},
			}
		}

		if _, ok := cache.InstanceData[instanceName]; !ok {
			cache.InstanceData[instanceName] = InstanceCache{}
		}
	}

	defer func() {
		if err := settingsManager.Save(cache, user, org); err != nil {
			log.Fatal(err)
			return
		}
	}()

	iconPath := "DefaultIcon.ico"
	if org.SystemTrayAppIcon != "" {
		iconPath = org.SystemTrayAppIcon
	}

	icon, _ := fyne.LoadResourceFromPath(iconPath)
	a := app.New()

	if desk, ok := a.(desktop.App); ok {
		// Compute a sorted list of all the available instances
		availableInstances := make([]string, 0, len(org.InstanceData))
		for k := range org.InstanceData {
			availableInstances = append(availableInstances, k)
		}
		sort.Strings(availableInstances)

		// Create the main menu
		var menu *fyne.Menu = fyne.NewMenu("GoTalk")

		// Create the various submenus
		for _, instance := range availableInstances {
			var showUserNotifications *fyne.MenuItem
			var showGroupNotifications *fyne.MenuItem
			var showBotNotifications *fyne.MenuItem
			var showGuestNotifications *fyne.MenuItem
			var showBridgedNotifications *fyne.MenuItem
			var showMutedNotifications *fyne.MenuItem
			var playNotificationSounds *fyne.MenuItem

			updateSettings := func() {
				data := user.InstanceData[instance]
				data.NotificationSettings.ShowUserNotifications = showUserNotifications.Checked
				data.NotificationSettings.ShowGroupNotifications = showGroupNotifications.Checked
				data.NotificationSettings.ShowBotNotifications = showBotNotifications.Checked
				data.NotificationSettings.ShowGuestNotifications = showGuestNotifications.Checked
				data.NotificationSettings.ShowBridgedNotifications = showBridgedNotifications.Checked
				data.NotificationSettings.ShowMutedNotifications = showMutedNotifications.Checked
				data.NotificationSettings.PlayNotificationSounds = playNotificationSounds.Checked
				user.InstanceData[instance] = data

				settingsManager.Save(cache, user, org)
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
			playNotificationSounds = fyne.NewMenuItem("Play Notification Sounds", func() {
				playNotificationSounds.Checked = !playNotificationSounds.Checked
				updateSettings()
				menu.Refresh()
			})

			data := user.InstanceData[instance]
			showUserNotifications.Checked = data.NotificationSettings.ShowUserNotifications
			showGroupNotifications.Checked = data.NotificationSettings.ShowGroupNotifications
			showBotNotifications.Checked = data.NotificationSettings.ShowBotNotifications
			showGuestNotifications.Checked = data.NotificationSettings.ShowGuestNotifications
			showBridgedNotifications.Checked = data.NotificationSettings.ShowBridgedNotifications
			showMutedNotifications.Checked = data.NotificationSettings.ShowMutedNotifications
			playNotificationSounds.Checked = data.NotificationSettings.PlayNotificationSounds

			submenu := fyne.NewMenuItem(instance, func() {})
			submenu.ChildMenu = fyne.NewMenu(
				instance,
				showUserNotifications,
				showGroupNotifications,
				showBotNotifications,
				showGuestNotifications,
				showBridgedNotifications,
				//fyne.NewMenuItemSeparator(),
				showMutedNotifications,
				playNotificationSounds,
			)

			coreItems := submenu.ChildMenu.Items
			loginItem := fyne.NewMenuItem("Log In", func() {})

			inst := org.InstanceData[instance]
			inst.setInstanceLoginMenuOption = func(callback func()) {
				if callback == nil {
					loginItem.Action = nil
					submenu.ChildMenu.Items = coreItems
					menu.Refresh()
				} else {
					submenu.ChildMenu.Items = coreItems
					loginItem.Action = callback
					submenu.ChildMenu.Items = append(submenu.ChildMenu.Items, loginItem)
					menu.Refresh()
				}
			}
			org.InstanceData[instance] = inst

			menu.Items = append(menu.Items, submenu)
		}

		var showNotifications *fyne.MenuItem
		var playNotificationSounds *fyne.MenuItem

		updateSettings := func() {
			user.ShowNotifications = showNotifications.Checked
			user.PlayNotificationSounds = playNotificationSounds.Checked

			settingsManager.Save(cache, user, org)
		}

		showNotifications = fyne.NewMenuItem("Show Notifications", func() {
			showNotifications.Checked = !showNotifications.Checked
			updateSettings()
			menu.Refresh()
		})

		playNotificationSounds = fyne.NewMenuItem("Play Notification Sounds", func() {
			playNotificationSounds.Checked = !playNotificationSounds.Checked
			updateSettings()
			menu.Refresh()
		})

		showNotifications.Checked = user.ShowNotifications
		playNotificationSounds.Checked = user.PlayNotificationSounds

		menu.Items = append(
			menu.Items,
			showNotifications,
			playNotificationSounds,
		)

		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayIcon(icon)
	}

	var wg sync.WaitGroup
	var closeChan chan interface{} = make(chan interface{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := startNextcloudMonitor(&wg, closeChan)
		if err != nil {
			log.Print(err)
			a.Quit()
		}
	}()

	defer systray.Quit()
	defer close(closeChan)

	a.Run()
}
