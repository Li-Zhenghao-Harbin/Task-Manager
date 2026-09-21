package main

import (
	"encoding/base64"
	"encoding/xml"
	"os"
	"os/exec"
	"strings"
	"unicode/utf16"
)

// toastAppID reuses PowerShell's own AppUserModelID. A toast needs a registered AppID to be
// displayed, and borrowing the shell's avoids making this app own a Start Menu shortcut.
const toastAppID = `{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe`

// toastScript renders a ToastGeneric notification from the XML handed over in TM_TOAST_XML.
// The script stays pure ASCII and reads its payload from the environment on purpose:
// PowerShell 5.1 decodes a BOM-less script file as ANSI, which would garble Chinese text if
// the message were embedded here. -EncodedCommand (UTF-16LE) sidesteps the same problem.
const toastScript = `$ErrorActionPreference = 'Stop'
$null = [Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime]
$null = [Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime]
$document = New-Object Windows.Data.Xml.Dom.XmlDocument
$document.LoadXml($env:TM_TOAST_XML)
$toast = New-Object Windows.UI.Notifications.ToastNotification $document
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($env:TM_TOAST_APPID).Show($toast)
`

// showWindowsToast displays a desktop notification through the WinRT toast API.
func showWindowsToast(title, body string) error {
	command := exec.Command("powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-WindowStyle", "Hidden",
		"-EncodedCommand", encodePowerShell(toastScript),
	)
	command.Env = append(os.Environ(),
		"TM_TOAST_APPID="+toastAppID,
		"TM_TOAST_XML="+toastXML(title, body),
	)
	return command.Run()
}

// toastXML builds a two-line ToastGeneric payload with the text properly escaped, so an Agent
// name containing & or < cannot break the document.
func toastXML(title, body string) string {
	var builder strings.Builder
	builder.WriteString(`<toast><visual><binding template="ToastGeneric"><text>`)
	builder.WriteString(xmlEscape(title))
	builder.WriteString(`</text><text>`)
	builder.WriteString(xmlEscape(body))
	builder.WriteString(`</text></binding></visual></toast>`)
	return builder.String()
}

func xmlEscape(value string) string {
	var builder strings.Builder
	_ = xml.EscapeText(&builder, []byte(value))
	return builder.String()
}

// encodePowerShell base64-encodes a script as UTF-16LE, the format -EncodedCommand expects.
func encodePowerShell(script string) string {
	units := utf16.Encode([]rune(script))
	raw := make([]byte, 0, len(units)*2)
	for _, unit := range units {
		raw = append(raw, byte(unit), byte(unit>>8))
	}
	return base64.StdEncoding.EncodeToString(raw)
}
