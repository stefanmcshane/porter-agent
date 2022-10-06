package lokistore

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/porter-dev/porter-agent/pkg/logstore"
	"github.com/porter-dev/porter-agent/pkg/logstore/lokistore/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/grafana/loki/pkg/logcli/client"
	"github.com/grafana/loki/pkg/loghttp"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/grafana/loki/pkg/logqlmodel"
)

type LokiStore struct {
	name          string
	address       string
	httpAddress   string
	pusherClient  proto.PusherClient
	querierClient proto.QuerierClient
	closer        io.Closer
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

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, fmt.Errorf("error initializing loki client with name %s. Error: %w", name, err)
	}

	return &LokiStore{
		address:       address,
		httpAddress:   config.HTTPAddress,
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
		Labels:  logstore.LabelsMapToString(labels, "=", ""),
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
	c := client.DefaultClient{
		Address: store.httpAddress,
	}

	qrResp, err := c.QueryRange(
		logstore.ConstructSearch(logstore.LabelsMapToString(options.Labels, "=~", options.CustomSelectorSuffix), options.SearchParam),
		int(options.Limit),
		options.End,
		options.Start,
		logproto.BACKWARD,
		time.Hour,
		time.Hour,
		false,
	)

	if err != nil {
		return err
	}

	switch qrResp.Data.Result.Type() {
	case logqlmodel.ValueTypeStreams:
		fmt.Println("IS A STREAM TYPE")
	case loghttp.ResultTypeMatrix:
		fmt.Println("IS A matrix TYPE")
	case loghttp.ResultTypeVector:
		fmt.Println("IS A VECTOR TYPE")
	default:
		fmt.Println("NONE OF THOSE TYPES")
	}

	streams := qrResp.Data.Result.(loghttp.Streams)

	for _, s := range streams {
		for _, entry := range s.Entries {
			err := w.Write(&entry.Timestamp, entry.Line)

			if err != nil {
				return err
			}
		}
	}

	return nil

	// stream, err := store.querierClient.Query(context.Background(), &proto.QueryRequest{
	// 	Selector:  logstore.ConstructSearch(logstore.LabelsMapToString(options.Labels, "=~", options.CustomSelectorSuffix), options.SearchParam),
	// 	Start:     timestamppb.New(options.Start),
	// 	End:       timestamppb.New(options.End),
	// 	Limit:     options.Limit,
	// 	Direction: proto.Direction_BACKWARD,
	// })

	// if err != nil {
	// 	return fmt.Errorf("error querying logs from loki store with name %s. Error: %w", store.name, err)
	// }

	// for _, val := range qrResp.Data.ResultType {

	// }

	// for {
	// 	select {
	// 	case <-stopCh:
	// 		return nil
	// 	default:
	// 		resp, err := stream.Recv()

	// 		if err != nil {
	// 			if err == io.EOF {
	// 				return nil
	// 			}

	// 			return err
	// 		}

	// 		for _, s := range resp.GetStreams() {
	// 			for _, entry := range s.GetEntries() {
	// 				t := entry.Timestamp.AsTime()

	// 				err := w.Write(&t, entry.Line)

	// 				if err != nil {
	// 					return err
	// 				}
	// 			}
	// 		}
	// 	}
	// }
}

func (store *LokiStore) Tail(options logstore.TailOptions, w logstore.Writer, stopCh <-chan struct{}) error {
	stream, err := store.querierClient.Tail(context.Background(), &proto.TailRequest{
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
	labelValues, err := store.querierClient.Label(context.Background(), &proto.LabelRequest{
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
