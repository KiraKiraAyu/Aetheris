package delivery

import (
	"bytes"
	"encoding/json"
	"text/template"

	"aetheris/internal/notification"
)

type templateData struct {
	ID             string
	TenantID       string
	Recipient      string
	Channel        string
	TemplateKey    string
	Title          string
	Body           string
	GroupKey       string
	AggregateCount int
	Metadata       notification.Metadata
	Text           string
}

func renderTemplate(name string, source string, record notification.Notification) (string, error) {
	tpl, err := template.New(name).
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"json":  jsonTemplateValue,
			"quote": jsonTemplateValue,
		}).
		Parse(source)
	if err != nil {
		return "", err
	}

	var rendered bytes.Buffer
	if err := tpl.Execute(&rendered, newTemplateData(record)); err != nil {
		return "", err
	}
	return rendered.String(), nil
}

func jsonTemplateValue(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func newTemplateData(record notification.Notification) templateData {
	return templateData{
		ID:             record.ID,
		TenantID:       record.TenantID,
		Recipient:      record.Recipient,
		Channel:        string(record.Channel),
		TemplateKey:    record.TemplateKey,
		Title:          record.Title,
		Body:           record.Body,
		GroupKey:       record.GroupKey,
		AggregateCount: record.AggregateCount,
		Metadata:       record.Metadata.Clone(),
		Text:           notificationText(record),
	}
}

func notificationText(record notification.Notification) string {
	switch {
	case record.Title != "" && record.Body != "":
		return record.Title + "\n" + record.Body
	case record.Title != "":
		return record.Title
	default:
		return record.Body
	}
}
