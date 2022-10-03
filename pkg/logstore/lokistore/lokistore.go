package lokistore

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/porter-dev/porter-agent/pkg/logstore"
	"github.com/porter-dev/porter-agent/pkg/logstore/lokistore/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LokiStore struct {
	name          string
	address       string
	pusherClient  proto.PusherClient
	querierClient proto.QuerierClient
	closer        io.Closer
}

type LogStoreConfig struct {
	Address string
}

func New(name string, config LogStoreConfig) (*LokiStore, error) {
	address := config.Address

	if address == "" {
		address = ":3100"
	}

	conn, err := grpc.Dial(address, grpc.WithInsecure())

	if err != nil {
		return nil, fmt.Errorf("error initializing loki client with name %s. Error: %w", name, err)
	}

	return &LokiStore{
		address:       address,
		name:          name,
		pusherClient:  proto.NewPusherClient(conn),
		querierClient: proto.NewQuerierClient(conn),
		closer:        conn,
	}, nil
}

func (store *LokiStore) Push(labels map[string]string, line string, t time.Time) error {
	entry := &proto.EntryAdapter{
		Timestamp: timestamppb.New(t),
		Line:      line,
	}

	streamAdapter := &proto.StreamAdapter{
		Labels:  logstore.LabelsMapToString(labels, "="),
		Entries: []*proto.EntryAdapter{entry},
	}

	_, err := store.pusherClient.Push(context.Background(), &proto.PushRequest{
		Streams: []*proto.StreamAdapter{streamAdapter},
	})

	if err != nil {
		return fmt.Errorf("error pushing to loki client. Error %w", err)
	}

	return nil
}

func (store *LokiStore) Query(options logstore.QueryOptions, w logstore.Writer, stopCh <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	stream, err := store.querierClient.Query(ctx, &proto.QueryRequest{
		Selector: logstore.LabelsMapToString(options.Labels, "=~"),
		Start:    timestamppb.New(options.Start),
		End:      timestamppb.New(options.End),
		Limit:    options.Limit,
	})

	if err != nil {
		return fmt.Errorf("error querying logs from loki store with name %s. Error: %w", store.name, err)
	}

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				resp, err := stream.Recv()

				for _, s := range resp.GetStreams() {
					for _, entry := range s.GetEntries() {
						w.Write(entry.Line)
					}
				}

				if err == io.EOF {
					return
				}

				if err != nil {
					return
				}

			}
		}
	}(ctx)

	<-stopCh
	cancel()

	return nil
}

func (store *LokiStore) Tail(options logstore.TailOptions, w logstore.Writer, stopCh <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	stream, err := store.querierClient.Tail(ctx, &proto.TailRequest{
		Query: logstore.LabelsMapToString(options.Labels, "=~"),
		Start: timestamppb.New(options.Start),
		Limit: options.Limit,
	})

	if err != nil {
		return fmt.Errorf("error streaming logs from loki store with name %s. Error: %w", store.name, err)
	}

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				resp, err := stream.Recv()

				entries := resp.Stream.GetEntries()

				for _, entry := range entries {
					w.Write(entry.GetLine())
				}

				if err == io.EOF {
					return
				}

				if err != nil {
					return
				}

			}
		}
	}(ctx)

	<-stopCh
	cancel()

	return nil
}
