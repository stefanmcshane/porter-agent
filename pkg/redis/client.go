package redis

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
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

func (c *Client) limitNumberOfBuckets(ctx context.Context, pattern string, limit int) {
	keys := c.client.Keys(ctx, pattern).Val()

	sort.Strings(keys)

	if len(keys) > limit {
		c.client.Del(ctx, keys[:len(keys)-limit]...).Val()
	}
}

func (c *Client) AppendAndTrimDetails(ctx context.Context, resourceType models.EventResourceType, namespace, name string, details []string) error {
	key := fmt.Sprintf("%s:%s:%s:%d", resourceType, namespace, name, time.Now().Unix())
	_, err := c.client.LPush(ctx, key, details).Result()
	if err != nil {
		return err
	}

	_, err = c.client.LTrim(ctx, key, 0, c.maxEntries).Result()
	if err != nil {
		return err
	}

	// set max TTL to 1 week
	_, err = c.client.Expire(ctx, key, 24*7*time.Hour).Result()
	if err != nil {
		return err
	}

	// limit max number of buckets to 20
	c.limitNumberOfBuckets(ctx, fmt.Sprintf("%s:%s:%s:*", resourceType, namespace, name), 20)

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

// SearchBestMatchForBucket first tries an exact match to return
// else resourts to matching the closest match for the given timestamp
func (c *Client) SearchBestMatchForBucket(ctx context.Context, resourceType models.EventResourceType, namespace, name, timestamp string) ([]string, string, error) {
	// see if passed in value is even a valid timestamp
	_, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return []string{}, "", err
	}

	// try exact match
	key := fmt.Sprintf("%s:%s:%s:%s", resourceType, namespace, name, timestamp)
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return []string{}, "", err
	}

	if exists == 1 {
		// exact match
		match, err := c.client.LRange(ctx, key, 0, -1).Result()
		return match, timestamp, err
	}

	// else get list of keys
	pattern := fmt.Sprintf("%s:%s:%s:*", resourceType, namespace, name)
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return []string{}, "", err
	}

	oldest := "0"

	for _, k := range keys {
		splits := strings.Split(k, ":")
		ts := splits[len(splits)-1]

		//fmt.Println("comparing", ts, key)

		if ts <= timestamp {
			if ts >= oldest {
				oldest = ts
			}
		}
	}

	if oldest == "0" {
		return []string{}, "", fmt.Errorf("cannot find a match")
	}

	// match for the key has been found, return the contents for that key
	matchPattern := fmt.Sprintf("%s:%s:%s:%s", resourceType, namespace, name, oldest)
	match, err := c.client.LRange(ctx, matchPattern, 0, -1).Result()

	return match, oldest, err
}
