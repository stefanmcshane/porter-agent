package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// logs are of the form "log:<incident_id>:<timestamp>"
type Log struct {
	incident  *Incident
	timestamp int64
}

func NewLog(incident *Incident, timestamp int64) *Log {
	return &Log{
		incident:  incident,
		timestamp: timestamp,
	}
}

func NewLogFromString(id string) (*Log, error) {
	segments := strings.Split(id, ":")

	if len(segments) != 6 || (len(segments) > 0 && segments[0] != "log") {
		return nil, fmt.Errorf("invalid log of the form: %s", id)
	}

	incident, err := NewIncidentFromString(strings.Join(segments[1:5], ":"))

	if err != nil {
		return nil, err
	}

	timestamp, err := strconv.ParseInt(segments[5], 10, 64)

	if err != nil {
		return nil, fmt.Errorf("error converting log timestamp to int: %w", err)
	}

	return &Log{
		incident:  incident,
		timestamp: timestamp,
	}, nil
}

func (l *Log) GetIncident() *Incident {
	return l.incident
}

func (l *Log) GetTimestamp() int64 {
	return l.timestamp
}

func (l *Log) GetTimestampAsTime() time.Time {
	return time.Unix(l.timestamp, 0)
}
