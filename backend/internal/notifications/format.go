package notifications

import "fmt"

// SeverityEmoji returns an emoji indicator for the given severity level.
func SeverityEmoji(severity string) string {
	switch severity {
	case "critical":
		return "\U0001F534" // red circle
	case "error":
		return "\U0001F7E0" // orange circle
	case "warning":
		return "\U0001F7E1" // yellow circle
	case "info":
		return "\U0001F535" // blue circle
	default:
		return "\u26AA" // white circle
	}
}

// SeverityColor returns a hex color code for Discord embeds.
func SeverityColor(severity string) int {
	switch severity {
	case "critical":
		return 0xED4245 // red
	case "error":
		return 0xE67E22 // orange
	case "warning":
		return 0xFEE75C // yellow
	case "info":
		return 0x3498DB // blue
	default:
		return 0x95A5A6 // grey
	}
}

// FormatSubject returns a notification subject line.
func FormatSubject(severity, appName string) string {
	return fmt.Sprintf("[Heimdall] %s: %s", severity, appName)
}

// FormatEmailHTML returns a minimal HTML email body.
func FormatEmailHTML(p Payload) string {
	return fmt.Sprintf(`<div style="font-family: monospace; max-width: 600px;">
<h2 style="margin: 0;">%s %s</h2>
<p style="color: #666; margin: 4px 0;">%s &mdash; %s</p>
<hr style="border: 1px solid #ddd;">
<p><strong>Summary:</strong> %s</p>
<div style="background: #f5f5f5; padding: 12px; border-radius: 4px; white-space: pre-wrap;">%s</div>
<hr style="border: 1px solid #ddd;">
<p style="color: #999; font-size: 12px;">Sent by Heimdall at %s</p>
</div>`, SeverityEmoji(p.Severity), FormatSubject(p.Severity, p.AppName), p.AppName, p.Severity, p.Summary, p.Assessment, p.Timestamp)
}
