package notify

import (
	"encoding/base64"
	"encoding/xml"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestBuildXMLEscapesContent(t *testing.T) {
	n := Notification{
		Title:               `"@ $(calc.exe) ]]><text>injected</text>`,
		Message:             "line 1\n\"@\n`$env:USERNAME` & <b>",
		Icon:                `C:\Users\a"b$\icon.png`,
		ActivationArguments: `https://example.org/call/a"b`,
		Actions:             []Action{{Label: "Open", Arguments: `https://example.org/?a=1&b=2`}},
	}

	data, err := n.buildXML()
	if err != nil {
		t.Fatal(err)
	}

	var parsed toastXML
	if err := xml.Unmarshal([]byte(data), &parsed); err != nil {
		t.Fatalf("invalid XML %q: %v", data, err)
	}

	if len(parsed.Binding.Texts) != 2 || parsed.Binding.Texts[0] != n.Title || parsed.Binding.Texts[1] != n.Message {
		t.Errorf("texts did not round-trip: %q", parsed.Binding.Texts)
	}
	if parsed.Binding.Image == nil || parsed.Binding.Image.Src != n.Icon {
		t.Errorf("icon did not round-trip: %+v", parsed.Binding.Image)
	}
	if parsed.Launch != n.ActivationArguments {
		t.Errorf("launch did not round-trip: %q", parsed.Launch)
	}
	if parsed.Binding.Template != "ToastGeneric" {
		t.Errorf("unexpected template %q", parsed.Binding.Template)
	}
	if parsed.Actions == nil || len(parsed.Actions.Actions) != 1 || parsed.Actions.Actions[0].Arguments != n.Actions[0].Arguments {
		t.Errorf("actions did not round-trip: %+v", parsed.Actions)
	}
	if parsed.Audio.Src != AudioDefault {
		t.Errorf("unexpected audio: %+v", parsed.Audio)
	}
}

func TestBuildXMLSilentWithoutActions(t *testing.T) {
	n := Notification{Title: "Title", Silent: true}

	data, err := n.buildXML()
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(data, "<actions") || strings.Contains(data, "<image") {
		t.Errorf("unexpected elements in %q", data)
	}
	if !strings.Contains(data, `<audio silent="true"></audio>`) {
		t.Errorf("missing silent audio in %q", data)
	}
}

func TestTruncate(t *testing.T) {
	long := strings.Repeat("è", maxTextLength+10)
	if got := []rune(truncate(long)); len(got) != maxTextLength+1 {
		t.Errorf("unexpected length %d", len(got))
	}
	if truncate("short") != "short" {
		t.Error("short strings must not change")
	}
}

func TestEncodeCommand(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString(encodeCommand("Write-Host 'è'"))
	if err != nil {
		t.Fatal(err)
	}

	codes := make([]uint16, len(data)/2)
	for i := range codes {
		codes[i] = uint16(data[2*i]) | uint16(data[2*i+1])<<8
	}
	if got := string(utf16.Decode(codes)); got != "Write-Host 'è'" {
		t.Errorf("unexpected decoded command %q", got)
	}
}
