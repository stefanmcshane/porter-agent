package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	porterErrors "github.com/porter-dev/porter-agent/pkg/errors"
	"github.com/porter-dev/porter-agent/pkg/models"
	"github.com/porter-dev/porter-agent/pkg/utils"
)

const (
	PODSTORE = iota
)

var agentCreationTimestamp int64 = 0

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

func (c *Client) IsFirstRun(ctx context.Context) (bool, error) {
	key := "porter-agent-creation-timestamp"

	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if exists == 1 {
		return true, nil
	}

	return false, nil
}

func (c *Client) SetAgentCreationTimestamp(ctx context.Context) error {
	exists, err := c.IsFirstRun(ctx)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("agent timestamp already exists in Redis")
	}

	key := "porter-agent-creation-timestamp"

	_, err = c.client.Set(ctx, key, time.Now().Unix(), 0).Result()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetAgentCreationTimestamp(ctx context.Context) (int64, error) {
	if agentCreationTimestamp != 0 {
		return agentCreationTimestamp, nil
	}

	key := "porter-agent-creation-timestamp"

	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	timestamp, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}

	agentCreationTimestamp = timestamp

	return agentCreationTimestamp, nil
}

func (c *Client) IncidentExists(ctx context.Context, incident string) (bool, error) {
	val, err := c.client.Exists(ctx, incident).Result()
	if err != nil {
		return false, fmt.Errorf("error checking if incident with ID: %s exists. Error: %w", incident, err)
	}

	if val == 1 {
		return true, nil
	} else if val > 1 {
		panic(fmt.Sprintf("more than one incident with key %s exists in Redis", incident))
	}

	return false, nil
}

func (c *Client) GetLatestEventForIncident(ctx context.Context, incidentID string) (*models.PodEvent, error) {
	data, err := c.client.ZRangeArgsWithScores(ctx, goredis.ZRangeArgs{
		Key:     incidentID,
		Start:   0,
		Stop:    -1,
		ByScore: true,
		Rev:     true,
		Count:   1,
	}).Result()
	if err != nil {
		return nil, err
	}

	payload, ok := data[0].Member.(string)
	if !ok {
		return nil, fmt.Errorf("error casting Redis Z Member to bytearray for incident ID: %s with score: %f",
			incidentID, data[0].Score)
	}

	event := &models.PodEvent{}

	err = json.Unmarshal([]byte(payload), event)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling event to json for incident ID: %s with score: %f. Error: %w",
			incidentID, data[0].Score, err)
	}

	return event, nil
}

func (c *Client) AddEventToIncident(ctx context.Context, incidentID string, event *models.PodEvent) error {
	// first check if incident is already in Redis
	newIncident, err := c.IncidentExists(ctx, incidentID)
	if err != nil {
		return err
	}

	if !newIncident {
		events, err := c.client.ZRange(ctx, incidentID, 0, -1).Result()
		if err != nil {
			return err
		}

		if len(events) >= 500 {
			// FIXME: a better way to do this?
			return fmt.Errorf("reached max event count of 500 for incident ID: %s", incidentID)
		}
	}

	score := time.Now().Unix()

	event.EventID = fmt.Sprintf("%s:%d", incidentID, score)

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshalling to JSON with event ID: %s. Error: %w", event.EventID, err)
	}

	_, err = c.client.ZAdd(ctx, incidentID, &goredis.Z{
		Score:  float64(score),
		Member: eventJSON,
	}).Result()
	if err != nil {
		return fmt.Errorf("error adding new pod event to incident with ID: %s. Error: %w", incidentID, err)
	}

	incidentObj, _ := utils.NewIncidentFromString(incidentID)

	if newIncident {
		// set a TTL for 2 weeks
		_, err = c.client.ExpireAt(ctx, incidentID, incidentObj.GetTimestampAsTime().Add(time.Hour*24*14)).Result()
		if err != nil {
			return fmt.Errorf("error setting expiration to incident with ID: %s. Error: %w", incidentID, err)
		}

		// we need to add this new incident to the pending queue so that it gets pushed out as a notification
		c.AppendToNotifyWorkQueue(ctx, []byte(incidentID))
	}

	_, err = c.client.SAdd(ctx, fmt.Sprintf("pods:%s", incidentID), event.PodName).Result()
	if err != nil {
		return fmt.Errorf("error adding new pod: %s to pod set with incident ID: %s. Error: %w",
			event.PodName, incidentID, err)
	}

	if newIncident {
		_, err = c.client.ExpireAt(ctx, fmt.Sprintf("pods:%s", incidentID),
			incidentObj.GetTimestampAsTime().Add(time.Hour*24*14)).Result()
		if err != nil {
			return fmt.Errorf("error setting expiration for pod set for incident ID: %s. Error: %w", incidentID, err)
		}
	}

	return nil
}

func (c *Client) SetPodResolved(ctx context.Context, podName, incidentID string) error {
	if exists, err := c.IncidentExists(ctx, incidentID); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("trying to set pod resolved for non-existent incident with ID: %s", incidentID)
	}

	key := fmt.Sprintf("pods:%s", incidentID)

	_, err := c.client.SRem(ctx, key, podName).Result()
	if err != nil {
		return fmt.Errorf("error trying to set pod resolved for pod: %s for incident ID: %s", podName, incidentID)
	}

	if affectedPods, err := c.client.SMembers(ctx, key).Result(); err != nil {
		return fmt.Errorf("error trying to get members of set: %s. Error: %w", key, err)
	} else if len(affectedPods) == 0 {
		// all pods are now healthy, delete the active incident

		incidentObj, _ := utils.NewIncidentFromString(incidentID)

		_, err := c.client.Del(ctx, fmt.Sprintf("active_incident:%s:%s", incidentObj.GetReleaseName(),
			incidentObj.GetNamespace())).Result()
		if err != nil {
			return fmt.Errorf("error trying to remove %s from active_incident. Error: %w", incidentID, err)
		}
	}

	return nil
}

func (c *Client) IsIncidentResolved(ctx context.Context, incidentID string) (bool, error) {
	pods, err := c.client.SMembers(ctx, fmt.Sprintf("pods:%s", incidentID)).Result()
	if err != nil {
		return false, fmt.Errorf("error getting pod members for incident ID: %s. Error: %w", incidentID, err)
	}

	if len(pods) == 0 {
		return true, nil
	}

	return false, nil
}

func (c *Client) GetAllIncidents(ctx context.Context) ([]string, error) {
	incidents, err := c.client.Keys(ctx, "incident:*:*:*").Result()
	if err != nil {
		return nil, err
	}

	return incidents, nil
}

func (c *Client) GetIncidentsByReleaseNamespace(ctx context.Context, releaseName, namespace string) ([]string, error) {
	incidents, err := c.client.Keys(ctx, fmt.Sprintf("incident:%s:%s:*", releaseName, namespace)).Result()
	if err != nil {
		return nil, err
	}

	return incidents, nil
}

func (c *Client) GetIncidentEventsByID(ctx context.Context, incidentID string) ([]*models.PodEvent, error) {
	payload, err := c.client.ZRangeArgsWithScores(ctx, goredis.ZRangeArgs{
		Key:     incidentID,
		Start:   0,
		Stop:    -1,
		ByScore: true,
		Rev:     true,
		Count:   500, // let us enforce a total of 500 pod events here, FIXME: this should happen while adding new events
	}).Result()
	if err != nil {
		return nil, err
	}

	var events []*models.PodEvent

	for _, data := range payload {
		event := &models.PodEvent{}

		rawBytes, ok := data.Member.(string)
		if !ok {
			return nil, fmt.Errorf("error casting Redis Z Member to bytearray for incident ID: %s with score: %f",
				incidentID, data.Score)
		}

		err = json.Unmarshal([]byte(rawBytes), event)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling event to json for incident ID: %s with score: %f. Error: %w",
				incidentID, data.Score, err)
		}

		events = append(events, event)
	}

	return events, nil
}

func (c *Client) AddLogs(ctx context.Context, strLogs string) (string, error) {
	logID := uuid.New().String()

	if _, err := c.client.Set(ctx, fmt.Sprintf("log:%s", logID), strLogs, time.Hour*24*14).Result(); err != nil {
		return "", errors.New("error adding logs")
	}

	return logID, nil
}

func (c *Client) GetLogs(ctx context.Context, logID string) (string, error) {
	if exists, err := c.client.Exists(ctx, logID).Result(); err != nil {
		return "", fmt.Errorf("error fetching logs with ID: %s", logID)
	} else if exists == 0 {
		return "", fmt.Errorf("no such logs with ID: %s", logID)
	}

	logs, err := c.client.Get(ctx, logID).Result()
	if err != nil {
		return "", fmt.Errorf("error fetching logs with ID: %s. Error: %w", logID, err)
	}

	return logs, nil
}

func (c *Client) GetLatestReasonAndMessage(ctx context.Context, incidentID string) (string, string, error) {
	if exists, err := c.IncidentExists(ctx, incidentID); err != nil {
		return "", "", fmt.Errorf("error checking for existence of incident with ID: %s. Error: %w", incidentID, err)
	} else if !exists {
		return "", "", fmt.Errorf("trying to fetch reason, message for non-existent incident with ID: %s", incidentID)
	}

	event, err := c.GetLatestEventForIncident(ctx, incidentID)
	if err != nil {
		return "", "", err
	}

	return event.Reason, event.Message, nil
}

func (c *Client) GetOrCreateActiveIncident(ctx context.Context, releaseName, namespace string) (string, error) {
	key := fmt.Sprintf("active_incident:%s:%s", releaseName, namespace)

	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("error checking for active incident for %s in namespace %s. Error: %w",
			releaseName, namespace, err)
	}

	if exists == 0 {
		// create a new active incident key
		newIncident := utils.NewIncident(releaseName, namespace, time.Now().Unix())

		_, err := c.client.Set(ctx, key, newIncident.ToString(), time.Hour*24*14).Result()
		if err != nil {
			return "", fmt.Errorf("error creating new active incident for release %s with namespace %s. Error: %w",
				releaseName, namespace, err)
		}
	} else if exists == 1 {
		incidentID, err := c.client.Get(ctx, key).Result()
		if err != nil {
			return "", fmt.Errorf("error fetching active incident for %s in namespace %s. Error: %w",
				releaseName, namespace, err)
		}

		return incidentID, nil
	}

	return "", fmt.Errorf("error fetching active incident for %s in namespace %s", releaseName, namespace)
}
