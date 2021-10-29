package redis

import (
	"context"
	"testing"

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
