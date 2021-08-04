package redis

import (
	"context"
	"fmt"

	redis "github.com/go-redis/redis/v8"
)

const (
	PODSTORE = iota
	HPASTORE
	NODESTORE
)

type Client struct {
	client     *redis.Client
	maxEntries int64
}

func NewClient(host, port, username, password string, db int, maxEntries int64) *Client {
	return &Client{
		client: redis.NewClient(&redis.Options{
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
