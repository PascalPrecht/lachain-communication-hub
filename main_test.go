package main_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/grpc"
	"io"
	"lachain-communication-hub/config"
	server "lachain-communication-hub/grpc"
	"lachain-communication-hub/peer"
	"lachain-communication-hub/utils"
	"log"
	"testing"
	"time"

	pb "lachain-communication-hub/grpc/protobuf"
)

const (
	address1 = "localhost" + config.GRPCPort
	address2 = "localhost:50002"
)

func TestCommunication(t *testing.T) {
	// connect clients
	conn1, _ := makeServerPeer("_h1", config.GRPCPort, address1)
	defer conn1.Close()

	conn2, pub := makeServerPeer("_h2", ":50002", address2)
	defer conn2.Close()

	client := pb.NewCommunicationHubClient(conn1)
	stream, err := client.Communicate(context.Background())
	if err != nil {
		log.Fatalf("openn stream error %v", err)
	}
	done := make(chan bool)

	go func() {
		req := pb.InboundMessage{
			PublicKey: pub,
			Data:      []byte("ping"),
		}
		if err := stream.Send(&req); err != nil {
			log.Fatalf("can not send %v", err)
		}
	}()

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				fmt.Println("EOF")
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			log.Println("received grpc message:", string(resp.Data))
			stream.CloseSend()
		}
	}()

	<-done
}

func makeServerPeer(id string, port string, address string) (*grpc.ClientConn, []byte) {
	p := peer.New(id)
	server.New(port, &p)

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	c := pb.NewCommunicationHubClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	getKeyResult, err := c.GetKey(ctx, &pb.GetHubIdRequest{})
	if err != nil {
		log.Fatalf("could not GetKey: %v", err)
	}
	cancel()

	prv, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("could not GenerateKey: %v", err)
	}
	pub := crypto.CompressPubkey(&prv.PublicKey)

	fmt.Println("pubKey", hex.EncodeToString(pub))

	signature, err := utils.LaSign(getKeyResult.Id, prv)
	if err != nil {
		panic(err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)

	initR, err := c.Init(ctx, &pb.InitRequest{
		Signature: signature,
	})
	if err != nil {
		log.Fatalf("could not Init: %v", err)
	}
	cancel()

	log.Println("init result:", initR.Result)
	return conn, pub
}