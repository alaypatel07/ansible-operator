package eventapi

import (
	"fmt"
	"strings"
	"time"
)

// EventTime - time to unmarshal nano time.
type EventTime struct {
	time.Time
}

// UnmarshalJSON - override unmarshal json.
func (e *EventTime) UnmarshalJSON(b []byte) (err error) {
	e.Time, err = time.Parse("2006-01-02T15:04:05.999999999", strings.Trim(string(b[:]), "\"\\"))
	if err != nil {
		return err
	}
	return nil
}

// MarshalJSON - override the marshal json.
func (e EventTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", e.Time.Format("2006-01-02T15:04:05.99999999"))), nil
}

// JobEvent - event of an ansible run.
type JobEvent struct {
	UUID      string                 `json:"uuid"`
	Counter   int                    `json:"counter"`
	StdOut    string                 `json:"stdout"`
	StartLine int                    `json:"start_line"`
	EndLine   int                    `json:"EndLine"`
	Event     string                 `json:"event"`
	EventData map[string]interface{} `json:"event_data"`
	PID       int                    `json:"pid"`
	Created   EventTime              `json:"created"`
}

// StatusJobEvent - event of an ansible run.
type StatusJobEvent struct {
	UUID      string         `json:"uuid"`
	Counter   int            `json:"counter"`
	StdOut    string         `json:"stdout"`
	StartLine int            `json:"start_line"`
	EndLine   int            `json:"EndLine"`
	Event     string         `json:"event"`
	EventData StatsEventData `json:"event_data"`
	PID       int            `json:"pid"`
	Created   EventTime      `json:"created"`
}

// StatsEventData - data for a the status event.
type StatsEventData struct {
	Playbook     string         `json:"playbook"`
	PlaybookUUID string         `json:"playbook_uuid"`
	Changed      map[string]int `json:"changed"`
	Ok           map[string]int `json:"ok"`
	Failures     map[string]int `json:"failures"`
	Skipped      map[string]int `json:"skipped"`
}
