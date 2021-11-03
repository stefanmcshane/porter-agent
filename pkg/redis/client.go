package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"
	porterErrors "github.com/porter-dev/porter-agent/pkg/errors"
	"github.com/porter-dev/porter-agent/pkg/models"
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

func (c *Client) AppendAndTrimDetails(ctx context.Context, resourceType, namespace, name string, details []string) error {
	key := fmt.Sprintf("%s:%s:%s:%d", resourceType, namespace, name, time.Now().Unix())
	_, err := c.client.LPush(ctx, key, details).Result()
	if err != nil {
		return err
	}

	_, err = c.client.LTrim(ctx, key, 0, c.maxEntries).Result()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetDetails(ctx context.Context, resourceType, namespace, name string) ([]string, error) {
	key := fmt.Sprintf("%s:%s:%s", resourceType, namespace, name)
	return c.client.LRange(ctx, key, 0, -1).Result()
}

func (c *Client) AppendToNotifyWorkQueue(ctx context.Context, packed []byte) error {
	key := "pending"

	_, err := c.client.ZAdd(ctx, key, &goredis.Z{
		Score:  float64(time.Now().Unix()),
		Member: packed,
	}).Result()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetItemFromPendingQueue(ctx context.Context) ([]byte, float64, error) {
	key := "pending"

	// check if there's any item in pending queue
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return []byte{}, 0, err
	}

	if count == 0 {
		return []byte{}, 0, porterErrors.NoPendingItemError
	}

	value, err := c.client.ZPopMin(ctx, key).Result()
	if err != nil {
		return []byte{}, 0, err
	}

	// cast the member to byte array which was originally stored in the array
	member := value[0].Member
	rawBytes, ok := member.(string)
	if !ok {
		return []byte{}, 0, fmt.Errorf("cannot caste item to bytearray, actual type: %T", member)
	}

	return []byte(rawBytes), value[0].Score, nil
}

func (c *Client) RequeueItemWithScore(ctx context.Context, packed []byte, score float64) error {
	key := "pending"

	_, err := c.client.ZAdd(ctx, key, &goredis.Z{
		Score:  score,
		Member: packed,
	}).Result()

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RegisterErroredItem(ctx context.Context, resourceType models.EventResourceType, namespace, name string) error {
	key := fmt.Sprintf("errors:%s:%s:%s", resourceType, namespace, name)

	err := c.client.Set(ctx, key, true, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ErroredItemExists(ctx context.Context, resourceType models.EventResourceType, namespace, name string) (bool, error) {
	key := fmt.Sprintf("errors:%s:%s:%s", resourceType, namespace, name)

	val, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if val > 0 {
		return true, nil
	}

	return false, nil

}

func (c *Client) DeleteErroredItem(ctx context.Context, resourceType models.EventResourceType, namespace, name string) error {
	key := fmt.Sprintf("errors:%s:%s:%s", resourceType, namespace, name)

	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetKeysForResource(ctx context.Context, resourceType models.EventResourceType, namespace, name string) ([]string, error) {
	pattern := fmt.Sprintf("%s:%s:%s:*", resourceType, namespace, name)

	return c.client.Keys(ctx, pattern).Result()
}
