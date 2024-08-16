package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/weaviate/weaviate/cluster/proto/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("addr", "", "the cluster raft address")
	nodeId := flag.String("node-id", "", "the node if to remove")
	flag.Parse()

	if *nodeId == "" {
		panic("node-id must not be empty")
	}

	dialOp, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("error dialing: %s", err))
	}
	cl := api.NewClusterServiceClient(dialOp)

	resp, err := cl.RemovePeer(context.Background(), &api.RemovePeerRequest{Id: *nodeId})
	if err != nil {
		panic(fmt.Sprintf("error removing peer: %s", err))
	}
	fmt.Println(resp)
}
