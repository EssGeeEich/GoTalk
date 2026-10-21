// Package notify shows Windows toast notifications.
//
// The notification is built as XML in Go and handed to PowerShell through an
// environment variable, so its content is never interpreted as code: titles
// and messages come from chat messages, which anyone can write.
package notify

import (
	"encoding/base64"
	"encoding/xml"
	"unicode/utf16"
)

const (
	AudioDefault = "ms-winsoundevent:Notification.Default"
	AudioIM      = "ms-winsoundevent:Notification.IM"
)

// Windows limits environment variables to 32767 characters: keep texts well below that.
const maxTextLength = 2000

const envXML = "GOTALK_TOAST_XML"
const envAppID = "GOTALK_TOAST_APPID"

// Fixed script: the notification data is only ever read from the environment.
const script = `$ErrorActionPreference = 'Stop'
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.UI.Notifications.ToastNotification, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml($env:` + envXML + `)
$toast = New-Object Windows.UI.Notifications.ToastNotification $xml
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($env:` + envAppID + `).Show($toast)
`

type Action struct {
	Label     string // Text of the button
	Arguments string // URI opened when the button is clicked
}

type Notification struct {
	AppID               string   // Name shown in the Action Center
	Title               string   // Heading of the notification
	Message             string   // Body of the notification
	Icon                string   // Full path to an image file
	ActivationArguments string   // URI opened when the notification is clicked
	Actions             []Action // Optional buttons
	Audio               string   // One of the Audio* constants
	Silent              bool     // Don't play any sound
}

type toastXML struct {
	XMLName        xml.Name    `xml:"toast"`
	ActivationType string      `xml:"activationType,attr"`
	Launch         string      `xml:"launch,attr,omitempty"`
	Duration       string      `xml:"duration,attr"`
	Binding        bindingXML  `xml:"visual>binding"`
	Audio          audioXML    `xml:"audio"`
	Actions        *actionsXML `xml:"actions,omitempty"`
}

type bindingXML struct {
	Template string    `xml:"template,attr"`
	Image    *imageXML `xml:"image,omitempty"`
	Texts    []string  `xml:"text"`
}

type imageXML struct {
	Placement string `xml:"placement,attr"`
	Src       string `xml:"src,attr"`
}

type audioXML struct {
	Src    string `xml:"src,attr,omitempty"`
	Silent string `xml:"silent,attr,omitempty"`
}

type actionsXML struct {
	Actions []actionXML `xml:"action"`
}

type actionXML struct {
	ActivationType string `xml:"activationType,attr"`
	Content        string `xml:"content,attr"`
	Arguments      string `xml:"arguments,attr"`
}

func truncate(s string) string {
	runes := []rune(s)
	if len(runes) <= maxTextLength {
		return s
	}
	return string(runes[:maxTextLength]) + "…"
}

func (n *Notification) buildXML() (string, error) {
	t := toastXML{
		ActivationType: "protocol",
		Launch:         n.ActivationArguments,
		Duration:       "short",
		Binding:        bindingXML{Template: "ToastGeneric"},
	}

	// The binding element is only valid if it contains at least one text.
	t.Binding.Texts = append(t.Binding.Texts, truncate(n.Title))
	if n.Message != "" {
		t.Binding.Texts = append(t.Binding.Texts, truncate(n.Message))
	}

	if n.Icon != "" {
		t.Binding.Image = &imageXML{Placement: "appLogoOverride", Src: n.Icon}
	}

	if n.Silent {
		t.Audio.Silent = "true"
	} else if n.Audio != "" {
		t.Audio.Src = n.Audio
	} else {
		t.Audio.Src = AudioDefault
	}

	if len(n.Actions) > 0 {
		t.Actions = &actionsXML{}
	}
	for _, a := range n.Actions {
		t.Actions.Actions = append(t.Actions.Actions, actionXML{
			ActivationType: "protocol",
			Content:        a.Label,
			Arguments:      a.Arguments,
		})
	}

	data, err := xml.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func encodeCommand(command string) string {
	codes := utf16.Encode([]rune(command))
	data := make([]byte, 0, len(codes)*2)
	for _, c := range codes {
		data = append(data, byte(c), byte(c>>8))
	}
	return base64.StdEncoding.EncodeToString(data)
}
