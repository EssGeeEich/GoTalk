package main

type StartupType int64

const (
	// Send one notification each hour to remind the user to log in
	NotificationLogin StartupType = iota

	// Send just one notification at startup to remind the user to log in
	SingleNotificationLogin

	// Open the login page right away
	ImmediateLogin

	// Add a right-click menu option for logging in
	RightClickMenuLogin
)

type UserInstanceSettings struct {
	Username             string // Username for logging in to Nextcloud
	EncryptedAppPassword string // AppPassword received through the Nextcloud Login Flow - Encrypted through DPAPI

	ShowUserNotifications    bool // Whether to show notifications for regular 1-on-1 chats
	ShowGroupNotifications   bool // Whether to show notifications for group chats or circles
	ShowBotNotifications     bool // Whether to show notifications for bot chats
	ShowGuestNotifications   bool // Whether to show notifications for guest chats
	ShowBridgedNotifications bool // Whether to show notifications for bridged chats
	ShowMutedNotifications   bool // Force notifications to be shown even if the conversation was muted
	PlayAudio                bool // Plays a notification sound
}

type UserSettings struct {
	ConnectedInstances map[string]UserInstanceSettings // Nextcloud instances that the user logged in to

	PlayAudio         bool // Global toggle for muting audio
	ShowNotifications bool // Global toggle for preventing notifications
}

type OrgInstanceSettings struct {
	InstanceURL string      // URL pointing to the Nextcloud instance
	Startup     StartupType // Chooses how to handle the application startup when the user isn't logged in yet
}

type OrgSettings struct {
	DefaultInstances map[string]OrgInstanceSettings // Nextcloud instances that the user will be prompted to login for
	CanAddInstances  bool                           // Determines whether the user can add other Nextcloud instances at will
	MessageCheckTime uint64                         // Time in seconds between each message notification check
}

func (u *UserSettings) BeforeSave() {

}

func (u *UserSettings) AfterSave() {

}

func (u *UserSettings) AfterLoad() error {
	return nil
}
