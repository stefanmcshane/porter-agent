package memorystore

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/nxadm/tail"
	"github.com/porter-dev/porter-agent/pkg/logstore"
)

type MemoryStore struct {
	name string
	t    *tail.Tail
}

func (store *MemoryStore) getLogFilePath() string {
	return path.Join("/var/tmp", store.name+".log")
}

func New(name string) (*MemoryStore, error) {
	store := new(MemoryStore)
	store.name = name

	logFilePath := store.getLogFilePath()
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
	logFilePath := store.getLogFilePath()

	f, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)

	if err != nil {
		return fmt.Errorf("error pushing log to memory store with name %s. Error: %w", store.name, err)
	}

	defer f.Close()

	if _, err := f.WriteString("\n" + log); err != nil {
		return fmt.Errorf("error pushing log to memory store with name %s. Error: %w", store.name, err)
	}

	return nil
}
