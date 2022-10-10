package lokistore

import (
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Reachability string

const (
	ReachableStatus   Reachability = "reachable"
	UnreachableStatus Reachability = "unreachable"
)

var (
	singleton   Reachability
	singletonLk sync.Mutex

	lokiAddress string
)

func SetupLokiStatus(addr string) {
	lokiAddress = addr

	go func() {
		for {
			updateLokiStatus()

			time.Sleep(5 * time.Second)
		}
	}()
}

func GetLokiStatus() Reachability {
	singletonLk.Lock()
	defer singletonLk.Unlock()

	if singleton == "" {
		updateLokiStatus()
	}

	return singleton
}

func updateLokiStatus() {
	singletonLk.Lock()
	defer singletonLk.Unlock()

	if lokiAddress == "" {
		singleton = UnreachableStatus
		return
	}

	conn, err := grpc.Dial(lokiAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))

	defer closeConnIfExists(conn)

	if err != nil {
		// TODO: we can use grpc.Code() to get the exact error codes
		//       as defined in https://pkg.go.dev/google.golang.org/grpc/codes
		singleton = UnreachableStatus
	} else {
		singleton = ReachableStatus
	}
}

func closeConnIfExists(conn *grpc.ClientConn) {
	if conn != nil {
		conn.Close()
	}
}
