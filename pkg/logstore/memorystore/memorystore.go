package memorystore

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/nxadm/tail"
	"github.com/porter-dev/porter-agent/pkg/logstore"
)

type MemoryStore struct {
	name     string
	location string
	t        *tail.Tail
}

type Options struct {
	Dir string // Store the log file at this location. Defaults to /var/tmp
}

func (store *MemoryStore) createLogFile() error {
	logFilePath := store.location

	logFileDir := filepath.Dir(logFilePath)

	err := os.MkdirAll(logFileDir, os.ModePerm)

	if err != nil {
		return fmt.Errorf("error creating log directory for memory store with name %s. Error: %w", store.name, err)
	}

	f, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		return fmt.Errorf("error creating log file for memory store with name %s. Error: %w", store.name, err)
	}

	defer f.Close()

	return nil
}

func New(name string, options Options) (*MemoryStore, error) {
	store := new(MemoryStore)
	store.name = name

	logFileDir := options.Dir

	if logFileDir == "" {
		logFileDir = "/var/tmp"
	}

	store.location = path.Join(logFileDir, name+".log")

	err := store.createLogFile()

	if err != nil {
		return nil, err
	}

	logFilePath := store.location
	t, err := tail.TailFile(logFilePath, tail.Config{Follow: true})

	if err != nil {
		return nil, fmt.Errorf("error initializing memory store with name %s. Error: %w", store.name, err)
	}

	store.t = t
	return store, nil
}

func (store *MemoryStore) Stream(w logstore.Writer) {
	for line := range store.t.Lines {
		if strings.TrimSpace(line.Text) == "" {
			continue
		}

		w.Write(line.Text)
	}
}

func (store *MemoryStore) Stop() error {
	store.t.Cleanup()
	err := store.t.Stop()

	if err != nil {
		return fmt.Errorf("error stopping memory stream with name %s. Error: %w", store.name, err)
	}

	return nil
}

func (store *MemoryStore) Push(log string) error {
	logFilePath := store.location

	f, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)

	if err != nil {
		return fmt.Errorf("error opening log file for memory store with name %s. Error: %w", store.name, err)
	}

	defer f.Close()

	if _, err := f.WriteString("\n" + log); err != nil {
		return fmt.Errorf("error pushing log to memory store with name %s. Error: %w", store.name, err)
	}

	return nil
}
