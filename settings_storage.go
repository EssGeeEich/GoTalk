package main

import "GoTalk/nc"

type StartupType int64

const (
	// Don't send a login prompt
	NoLoginPrompt StartupType = iota

	// Send one notification to remind the user to log in
	NotificationLogin

	// Open the login page right away
	ImmediateLogin
)

type InstanceCache struct {
	Username             string // Username for logging in to Nextcloud
	EncryptedAppPassword string // AppPassword received through the Nextcloud Login Flow - Encrypted through DPAPI
}

type Cache struct {
	InstanceData map[string]InstanceCache
}

type UserInstanceSettings struct {
	NotificationSettings nc.NotificationSettings
}

type UserSettings struct {
	InstanceData map[string]UserInstanceSettings // Nextcloud instances that the user logged in to

	PlayAudio         bool // Global toggle for muting audio
	ShowNotifications bool // Global toggle for preventing notifications
}

type OrgInstanceSettings struct {
	InstanceURL         string      // URL pointing to the Nextcloud instance
	Startup             StartupType // Chooses how to handle the application startup when the user isn't logged in yet
	NotificationAppIcon string      // Custom App Icon. Uses DefaultIcon.png otherwise. Should be a full path pointing to a PNG file.
}

type OrgSettings struct {
	InstanceData      map[string]OrgInstanceSettings // Nextcloud instances that the user will be prompted to login for
	MessageCheckTime  uint64                         // Time in seconds between each message notification check
	SystemTrayAppIcon string                         // Custom App Icon. Uses DefaultIcon.ico otherwise. Should be a full path pointing to a ICO file.
}
