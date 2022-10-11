package sendgridsender

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/gomarkdown/markdown"
)

const templateFilename = "emailTemplate.html"

func Template(title string, content []EmailNotificationsContent) (string, error) {
	tpl, err := readTemplateFile(templateFilename)
	if err != nil {
		return "", fmt.Errorf("email %q: %v", title, err)
	}
	t, err := template.New("email").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("template %q: %v", templateFilename, err)
	}

	data := struct {
		Title         string
		Notifications []EmailNotificationsContent
	}{
		Title:         title,
		Notifications: content,
	}

	for i, n := range data.Notifications {
		// TODO(lds): This may be unsecure on untrusted data. Generally data comes from the platform or our code.
		data.Notifications[i].Message = string(markdown.ToHTML([]byte(n.Message), nil, nil))
	}

	emailContent := strings.Builder{}
	if err := t.Execute(&emailContent, data); err != nil {
		return "", fmt.Errorf("template: %v", err)
	}
	return emailContent.String(), nil
}

func readTemplateFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("error while reading template: %v", err)
	}
	return string(content), nil
}

type EmailNotificationsContent struct {
	Topic   string
	Message string
}
