package repository

import (
	"testing"

	"time"

	"github.com/porter-dev/porter-agent/internal/models"
)

func TestCreateEventCache(t *testing.T) {
	tester := &tester{
		dbFileName: "./event_cache_test.db",
	}

	setupTestEnv(tester, t)
	defer cleanup(tester, t)

	timestamp := time.Now()

	evCache := &models.EventCache{
		EventUID:     "very-unique-id",
		PodName:      "mighty-pod",
		PodNamespace: "mighty-namespace",
		Timestamp:    &timestamp,
	}

	_, err := tester.repo.EventCache.CreateEventCache(evCache)

	if err != nil {
		t.Fatalf("Expected no error after creating event cache, got %v", err)
	}

	cacheEntries, err := tester.repo.EventCache.ListEventCachesForEvent("very-unique-id")

	if err != nil {
		t.Fatalf("Expected no error after listing event caches, got %v", err)
	}

	if len(cacheEntries) != 1 {
		t.Fatalf("Expected 1 cache entry, got %d", len(cacheEntries))
	}

	if cacheEntries[0].ID != 1 {
		t.Fatalf("Expected event cache ID to be 1, got %d", cacheEntries[0].ID)
	}

	if cacheEntries[0].EventUID != "very-unique-id" {
		t.Fatalf("Expected event uid to be 'very-unique-id', got '%s'", cacheEntries[0].EventUID)
	}

	if cacheEntries[0].PodName != "mighty-pod" {
		t.Fatalf("Expected pod name to be 'mighty-pod', got '%s'", cacheEntries[0].PodName)
	}

	if cacheEntries[0].PodNamespace != "mighty-namespace" {
		t.Fatalf("Expected pod namespace to be 'mighty-namespace', got '%s'", cacheEntries[0].PodNamespace)
	}

	if cacheEntries[0].Timestamp.Unix() != timestamp.Unix() {
		t.Fatalf("Expected timestamp to be '%v', got '%v'", timestamp, cacheEntries[0].Timestamp)
	}
}

func TestListEventCachesForEvent(t *testing.T) {
	tester := &tester{
		dbFileName: "./event_cache_test.db",
	}

	setupTestEnv(tester, t)
	defer cleanup(tester, t)

	eventUIDs := []string{
		"very-unique-id-1",
		"very-unique-id-2",
		"very-unique-id-3",
	}

	for i, count := range []int{1, 2, 3} {
		for j := 0; j < count; j++ {
			timestamp := time.Now()

			_, err := tester.repo.EventCache.CreateEventCache(&models.EventCache{
				EventUID:  eventUIDs[i],
				Timestamp: &timestamp,
			})

			if err != nil {
				t.Fatalf("Expected no error after creating event cache, got %v", err)
			}
		}
	}

	for i, uid := range eventUIDs {
		caches, err := tester.repo.EventCache.ListEventCachesForEvent(uid)

		if err != nil {
			t.Fatalf("Expected no error after listing event caches, got %v", err)
		}

		if len(caches) != i+1 {
			t.Fatalf("Expected %d caches for event uid '%s', got %d", i+1, uid, len(caches))
		}
	}
}

func TestListEventCachesForPod(t *testing.T) {
	tester := &tester{
		dbFileName: "./event_cache_test.db",
	}

	setupTestEnv(tester, t)
	defer cleanup(tester, t)

	timestamp := time.Now()

	evCache := &models.EventCache{
		EventUID:     "very-unique-id",
		PodName:      "mighty-pod",
		PodNamespace: "mighty-namespace",
		Timestamp:    &timestamp,
	}

	_, err := tester.repo.EventCache.CreateEventCache(evCache)

	if err != nil {
		t.Fatalf("Expected no error after creating event cache, got %v", err)
	}

	cacheEntries, err := tester.repo.EventCache.ListEventCachesForPod("mighty-pod", "mighty-namespace")

	if err != nil {
		t.Fatalf("Expected no error after listing event caches, got %v", err)
	}

	if len(cacheEntries) != 1 {
		t.Fatalf("Expected 1 cache entry, got %d", len(cacheEntries))
	}

	if cacheEntries[0].EventUID != "very-unique-id" {
		t.Fatalf("Expected event uid to be 'very-unique-id', got '%s'", cacheEntries[0].EventUID)
	}

	if cacheEntries[0].PodName != "mighty-pod" {
		t.Fatalf("Expected pod name to be 'mighty-pod', got '%s'", cacheEntries[0].PodName)
	}

	if cacheEntries[0].PodNamespace != "mighty-namespace" {
		t.Fatalf("Expected pod namespace to be 'mighty-namespace', got '%s'", cacheEntries[0].PodNamespace)
	}

	if cacheEntries[0].Timestamp.Unix() != timestamp.Unix() {
		t.Fatalf("Expected timestamp to be '%v', got '%v'", timestamp, cacheEntries[0].Timestamp)
	}
}
