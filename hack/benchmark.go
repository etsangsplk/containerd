// single app that will run containers in containerd and output
// the total time in seconds that it took for the execution.
// go run benchmark.go -count 1000 -bundle /containers/redis
package main

import (
	"flag"
	"strconv"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/containerd/api/grpc/types"
	netcontext "golang.org/x/net/context"
	"google.golang.org/grpc"
)

func init() {
	flag.StringVar(&bundle, "bundle", "/containers/redis", "the bundle path")
	flag.IntVar(&count, "count", 1000, "number of containers to run")
	flag.Parse()
}

var (
	count  int
	bundle string
	group  = sync.WaitGroup{}
	jobs   = make(chan string, 20)
)

func getClient() types.APIClient {
	conn, err := grpc.Dial("localhost:8888", grpc.WithInsecure())
	if err != nil {
		logrus.Fatal(err)
	}
	return types.NewAPIClient(conn)
}

func main() {
	client := getClient()
	for i := 0; i < 100; i++ {
		group.Add(1)
		go worker(client)
	}
	start := time.Now()
	for i := 0; i < count; i++ {
		id := strconv.Itoa(i)
		jobs <- id
	}
	close(jobs)
	group.Wait()
	end := time.Now()
	duration := end.Sub(start).Seconds()
	logrus.Info(duration)
}

func worker(client types.APIClient) {
	defer group.Done()
	for id := range jobs {
		if _, err := client.CreateContainer(netcontext.Background(), &types.CreateContainerRequest{
			Id:         id,
			BundlePath: bundle,
		}); err != nil {
			logrus.Error(err)
		}
	}
}