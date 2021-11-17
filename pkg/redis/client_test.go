package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/porter-dev/porter-agent/pkg/models"
)

func TestCheckErroredItem(t *testing.T) {
	client := NewClient("localhost", "6379", "", "", PODSTORE, 100)

	exists, err := client.ErroredItemExists(context.Background(), models.PodResource, "default", "ahins")
	if err != nil {
		t.Fatalf("error checking the errored items register. %s", err.Error())
	}

	if exists {
		t.Errorf("non existing item should result in false bool value. got %t", exists)
	}
}

func TestRegisterErroredItem(t *testing.T) {
	client := NewClient("localhost", "6379", "", "", PODSTORE, 100)

	err := client.RegisterErroredItem(context.Background(), models.PodResource, "default", "ahins")
	if err != nil {
		t.Fatalf("cannot add errored item to the register. %s", err.Error())
	}
}

func TestDeleteErroredItem(t *testing.T) {
	client := NewClient("localhost", "6379", "", "", PODSTORE, 100)

	err := client.DeleteErroredItem(context.Background(), models.PodResource, "default", "ahins")
	if err != nil {
		t.Fatalf("unable to delete the item from error register. %s", err.Error())
	}
}

func TestLimitNumberOfBuckets(t *testing.T) {
	client := NewClient("localhost", "6379", "", "", PODSTORE, 100)

	for i := 0; i < 25; i++ {
		err := client.AppendAndTrimDetails(context.Background(), models.PodResource, "default", "test", []string{"hello", "world"})
		if err != nil {
			t.Fatal("error from append and trim. error:", err.Error())
		}

		time.Sleep(1010 * time.Millisecond)
	}

	keys := client.client.Keys(context.Background(), fmt.Sprintf("%s:%s:%s*", models.PodResource, "default", "test")).Val()
	if len(keys) > 20 {
		t.Errorf("more than 20 keys. got: %d", len(keys))
	}
}
