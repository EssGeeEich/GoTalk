# GoTalk
A Nextcloud Talk client made in golang.


# What's this for
This client's only purpose is to send you a notification every time you receive a new message, and if you forget it'll send another notification after a while.

# Why's this for
I have personally found the official Nextcloud client to be lacking, occasionally leaving me with no notifications.
This client was built to improve the reliability of the Talk notifications.

# Who's this for
This client was made for S/M businesses looking to improve internal communications when Nextcloud is already set up.

Even the set-up process keeps this in mind.
There is no installer, you just copy the executable file wherever you deem worth, configure the Nextcloud instances with, say, a GPO, configure the automatic startup and you're done.

You can connect this tool with multiple Nextcloud instances at the same time.

# Configuration files
GoTalk manages three configuration files:

- Organization Configuration
- User Configuration
- User Cache

## Organization Configuration
This file is stored in `%APPDATA%/SGH/GoTalk/GoTalk.org.toml`

This file MUST be created and distributed from yourself/your administrator.

The file looks like the following:

```toml
# Message Check Time:
# Check for new messages every N seconds
MessageCheckTime = 5

# System Tray App Icon:
# Changes the default icon that's shown in the system tray
# Should be a full path pointing to a ICO file.
SystemTrayAppIcon = ''

[InstanceData]

# Here's the data for "My Nextcloud Instance"
[InstanceData.'My Nextcloud Instance']
# Instance URL:
# The Instance URL should follow this exact format:
InstanceURL = 'https://my-nextcloud-instance.example.org/'

# Login mode
# Chooses how to handle the application startup when the user isn't logged in yet.
# Must have one of the following values:
# 0 => Login with a Notification message
#  The login flow starts right away.
# The browser will be opened after the user clicks the notification message.
#  The user has about 20 minutes of time from when the notification pops up, after which the login flow fails.
# 1 => Login Immediately
#  The login flow starts right away.
#  The browser will open the login page immediately.
#  The user has about 20 minutes of time from when the browser is opened, after which the login flow fails.
# 2 => Login with Context Menu
#  The right-click menu will present a new "Login" option.
#  When clicked, the login flow will start.
#  The browser will open the login page.
#  The user has about 20 minutes to log into NextCloud since clicking the "Login" option.
Login = 1

# Notification Repeat Time:
# Defines after how many minutes a notification for the same chat should appear twice
# Minimum is 0.5 (30 seconds)
NotificationRepeatTime = 1.0

# Notification App Icon:
# Changes the default icon that pops up whenever a notification is received from this instance.
# Should be a full path pointing to a PNG file.
NotificationAppIcon = ''
```

## User Configuration
This file is stored in `%APPDATA%/SGH/GoTalk/GoTalk.user.toml`

This file will be automatically generated from GoTalk.
GoTalk can and will also change this file according to the user preferences.
The file looks like the following:

```toml
# Global toggle, set to false to hide all notifications
ShowNotifications = true
# Global toggle, set to false to disable all sounds
PlayNotificationSounds = true

[InstanceData]
[InstanceData.'My Nextcloud Instance']
[InstanceData.'My Nextcloud Instance'.NotificationSettings]
# Instance-specific toggles
# Show notifications from regular users
ShowUserNotifications = true
# Show notifications from groups
ShowGroupNotifications = true
# Show notifications from bots
ShowBotNotifications = true
# Show notifications from guests
ShowGuestNotifications = true
# Show notifications from bridged systems
ShowBridgedNotifications = true
# Ignore the "mute notifications" flag of the conversation
ShowMutedNotifications = false
# Play a notification sound
PlayNotificationSounds = true
```

## User Cache
This file is stored in `%LOCALAPPDATA%/SGH/GoTalk/GoTalk.cache.toml`

This file will be automatically generated from GoTalk.
GoTalk can and will also change this file according to the login status of the user.
This file looks like the following:

```toml
[InstanceData]
[InstanceData.'My Nextcloud Instance']
Username = 'my.nextcloud.username'
EncryptedAppPassword = 'base64-of-encrypted-password'
```

The password is encrypted using the Windows DPAPI -- specifically, CryptProtectData.
This means that the AppPassword can only be read back from a single user, from a specific machine.
This will ensure that a leak of this file will not immediately result in a security issue.