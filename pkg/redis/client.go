package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

const (
	PODSTORE = iota
	HPASTORE
	NODESTORE
)

// Client is a redis client that also holds the
// value for max log enteries to hold for each pod
type Client struct {
	client     *goredis.Client
	maxEntries int64
}

func NewClient(host, port, username, password string, db int, maxEntries int64) *Client {
	return &Client{
		client: goredis.NewClient(&goredis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Username: username,
			Password: password,
			DB:       db,
		}),
		maxEntries: maxEntries,
	}
}

func (r *Client) AppendAndTrimDetails(ctx context.Context, resourceType, namespace, name string, details []string) error {
	key := fmt.Sprintf("%s:%s:%s", resourceType, namespace, name)
	_, err := r.client.LPush(ctx, key, details).Result()
	if err != nil {
		return err
	}

	_, err = r.client.LTrim(ctx, key, 0, r.maxEntries).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r *Client) AppendToNotifyWorkQueue(ctx context.Context, resourceType, namespace, name string) error {
	key := "pending"

	value := fmt.Sprintf("%s:%s:%s", resourceType, namespace, name)
	_, err := r.client.ZAdd(ctx, key, &goredis.Z{
		Score:  float64(time.Now().Unix()),
		Member: value,
	}).Result()
	if err != nil {
		return err
	}

	return nil
}
