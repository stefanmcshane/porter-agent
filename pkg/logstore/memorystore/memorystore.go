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
		logFileDir = filepath.Join("var", "tmp")
	}

	store.location = path.Join(logFileDir, name+".log")

	err := store.createLogFile()

	if err != nil {
		return nil, err
	}

	return store, nil
}

func (store *MemoryStore) Stream(w logstore.Writer, stopCh <-chan struct{}) error {
	logFilePath := store.location

	t, err := tail.TailFile(logFilePath, tail.Config{Follow: true, Poll: true})

	if err != nil {
		return fmt.Errorf("error streaming memory store with name %s. Error: %w", store.name, err)
	}

	go func(t *tail.Tail) {
		for line := range t.Lines {
			if strings.TrimSpace(line.Text) == "" || line.Err != nil {
				continue
			}

			w.Write(line.Text)
		}
	}(t)

	<-stopCh
	t.Cleanup()
	t.Stop()

	return nil
}

func (store *MemoryStore) Push(log string) error {
	logFilePath := store.location

	f, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)

	if err != nil {
		return fmt.Errorf("error opening log file for memory store with name %s. Error: %w", store.name, err)
	}

	defer f.Close()

	if _, err := f.WriteString("\n" + log); err != nil {
		return fmt.Errorf("error pushing log to memory store with name %s. Error: %w", store.name, err)
	}

	return nil
}
