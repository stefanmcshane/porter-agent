package lokistore

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/porter-dev/porter-agent/pkg/logstore"
	"github.com/porter-dev/porter-agent/pkg/logstore/lokistore/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LokiStore struct {
	name    string
	address string
	client  *Client
}

type LogStoreConfig struct {
	Address     string
	HTTPAddress string
}

func New(name string, config LogStoreConfig) (*LokiStore, error) {
	address := config.Address

	if address == "" {
		address = ":3100"
	}

	client := NewClient(&LokiHTTPClientConf{
		Address: config.HTTPAddress,
	})

	fmt.Printf("creating new grpc connection to %s\n", address)

	return &LokiStore{
		address: address,
		name:    name,
		client:  client,
	}, nil
}

func (store *LokiStore) Push(labels map[string]string, line string, t time.Time) error {
	fmt.Printf("calling push on loki store\n")

	conn, err := grpc.Dial(store.address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return fmt.Errorf("error initializing loki client. Error: %w", err)
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	entry := &proto.EntryAdapter{
		Timestamp: timestamppb.New(t),
		Line:      line,
	}

	streamAdapter := &proto.StreamAdapter{
		Labels:  logstore.LabelsMapToString(labels, "=", ""),
		Entries: []*proto.EntryAdapter{entry},
	}

	_, err = proto.NewPusherClient(conn).Push(ctx, &proto.PushRequest{
		Streams: []*proto.StreamAdapter{streamAdapter},
	})

	if err != nil {
		return fmt.Errorf("error pushing to loki client. Error %w", err)
	}

	return nil
}

func (store *LokiStore) Query(options logstore.QueryOptions, w logstore.Writer, stopCh <-chan struct{}) error {
	fmt.Printf("calling query on loki store\n")

	conn, err := grpc.Dial(store.address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return fmt.Errorf("error initializing loki client. Error: %w", err)
	}

	defer conn.Close()

	qrResp, err := store.client.QueryRange(options)

	if err != nil {
		return err
	}

	for _, stream := range qrResp.Data.Result {
		for _, rawEntry := range stream.Values {
			if len(rawEntry) != 2 {
				continue
			}

			nano, err := strconv.ParseInt(rawEntry[0], 10, 64)

			if err != nil {
				continue
			}

			t := time.Unix(0, nano)

			err = w.Write(&t, rawEntry[1])

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (store *LokiStore) Tail(options logstore.TailOptions, w logstore.Writer, stopCh <-chan struct{}) error {
	fmt.Printf("calling tail on loki store\n")

	conn, err := grpc.Dial(store.address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return fmt.Errorf("error initializing loki client. Error: %w", err)
	}

	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	stream, err := proto.NewQuerierClient(conn).Tail(ctx, &proto.TailRequest{
		Query: logstore.ConstructSearch(logstore.LabelsMapToString(options.Labels, "=~", options.CustomSelectorSuffix), options.SearchParam),
		Start: timestamppb.New(options.Start),
		Limit: options.Limit,
	})

	if err != nil {
		return fmt.Errorf("error streaming logs from loki store with name %s. Error: %w", store.name, err)
	}

	for {
		select {
		case <-stopCh:
			return nil
		default:
			resp, err := stream.Recv()

			if err != nil {
				if err == io.EOF {
					return nil
				}

				return err
			}

			entries := resp.Stream.GetEntries()

			for _, entry := range entries {
				t := entry.Timestamp.AsTime()

				err := w.Write(&t, entry.Line)

				if err != nil {
					return err
				}
			}
		}
	}
}

func (store *LokiStore) GetLabelValues(options logstore.LabelValueOptions) ([]string, error) {
	fmt.Printf("calling get label values on loki store\n")

	conn, err := grpc.Dial(store.address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, fmt.Errorf("error initializing loki client. Error: %w", err)
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	labelValues, err := proto.NewQuerierClient(conn).Label(ctx, &proto.LabelRequest{
		Name:   options.Label,
		Values: true,
		Start:  timestamppb.New(options.Start),
		End:    timestamppb.New(options.End),
	})

	if err != nil {
		return nil, err
	}

	return labelValues.GetValues(), nil
}
