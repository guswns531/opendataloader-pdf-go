package sanitize

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

var (
	emailPattern = regexp.MustCompile(`(?i)\b[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}\b`)
	urlPattern   = regexp.MustCompile(`(?i)\b(?:https?|ftp)://[^\s<>()\[\]{}"']+|\bwww\.[^\s<>()\[\]{}"']+|\bmailto:[^\s<>()\[\]{}"']+`)
	ipv4Pattern  = regexp.MustCompile(`\b(?:25[0-5]|2[0-4]\d|1?\d?\d)(?:\.(?:25[0-5]|2[0-4]\d|1?\d?\d)){3}\b`)
	ipv6Pattern  = regexp.MustCompile(`(?i)\b[0-9a-f:]{2,}\b`)
	cardPattern  = regexp.MustCompile(`\b(?:\d[ -]*?){13,19}\b`)
	phonePattern = regexp.MustCompile(`(?:\+?\d[\d\s().-]{8,}\d)`)
)

const (
	emailPlaceholder = "[EMAIL]"
	urlPlaceholder   = "[URL]"
	ipPlaceholder    = "[IP]"
	cardPlaceholder  = "[CARD]"
	phonePlaceholder = "[PHONE]"
)

// Apply sanitizes sensitive strings in-place across the document graph.
func Apply(document *model.Document) error {
	if document == nil {
		return fmt.Errorf("sanitize filter requires a document")
	}

	for _, artifact := range document.Artifacts {
		sanitizeArtifact(artifact)
	}
	for _, page := range document.Pages {
		sanitizePage(page)
	}
	for _, element := range document.Kids {
		sanitizeContent(element)
	}

	return nil
}

func sanitizePage(page *model.Page) {
	if page == nil {
		return
	}

	for _, artifact := range page.Artifacts {
		sanitizeArtifact(artifact)
	}
	for _, element := range page.Kids {
		sanitizeContent(element)
	}
}

func sanitizeArtifact(artifact *model.RawArtifact) {
	if artifact == nil {
		return
	}

	artifact.Text = sanitizeText(artifact.Text)
	artifact.Style.Content = sanitizeText(artifact.Style.Content)
}

func sanitizeContent(element model.ContentElement) {
	if element == nil {
		return
	}

	switch node := element.(type) {
	case *model.Paragraph:
		node.TextProperties.Content = sanitizeText(node.TextProperties.Content)
	case *model.Heading:
		node.TextProperties.Content = sanitizeText(node.TextProperties.Content)
	case *model.Caption:
		node.TextProperties.Content = sanitizeText(node.TextProperties.Content)
	case *model.ListItem:
		node.TextProperties.Content = sanitizeText(node.TextProperties.Content)
		for _, kid := range node.Kids {
			sanitizeContent(kid)
		}
	case *model.TextBlock:
		for _, kid := range node.Kids {
			sanitizeContent(kid)
		}
	case *model.HeaderFooter:
		for _, kid := range node.Kids {
			sanitizeContent(kid)
		}
	case *model.TableCell:
		for _, kid := range node.Kids {
			sanitizeContent(kid)
		}
	case *model.Table:
		for _, row := range node.Rows {
			for _, cell := range row.Cells {
				sanitizeContent(cell)
			}
		}
	}
}

func sanitizeText(text string) string {
	if text == "" {
		return text
	}

	text = urlPattern.ReplaceAllStringFunc(text, func(match string) string {
		return urlPlaceholder
	})
	text = emailPattern.ReplaceAllString(text, emailPlaceholder)
	text = ipv4Pattern.ReplaceAllString(text, ipPlaceholder)
	text = ipv6Pattern.ReplaceAllStringFunc(text, func(match string) string {
		if looksLikeIPAddress(match) {
			return ipPlaceholder
		}
		return match
	})
	text = cardPattern.ReplaceAllStringFunc(text, func(match string) string {
		if looksLikeCardNumber(match) {
			return cardPlaceholder
		}
		return match
	})
	text = phonePattern.ReplaceAllStringFunc(text, func(match string) string {
		if looksLikePhoneNumber(match) {
			return phonePlaceholder
		}
		return match
	})

	return text
}

func looksLikeIPAddress(value string) bool {
	ip := net.ParseIP(strings.TrimSpace(value))
	return ip != nil
}

func looksLikeCardNumber(value string) bool {
	digits := digitsOnly(value)
	if len(digits) < 13 || len(digits) > 19 {
		return false
	}
	return luhnCheck(digits)
}

func looksLikePhoneNumber(value string) bool {
	digits := digitsOnly(value)
	if len(digits) < 10 || len(digits) > 15 {
		return false
	}
	compact := strings.TrimSpace(value)
	if compact == digits {
		return true
	}
	return strings.ContainsAny(compact, "()+-. ") || strings.HasPrefix(compact, "+")
}

func digitsOnly(value string) string {
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func luhnCheck(value string) bool {
	sum := 0
	double := false
	for i := len(value) - 1; i >= 0; i-- {
		digit := int(value[i] - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}
	return sum%10 == 0
}
