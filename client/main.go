package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"strings"
	"time"

	pb "anuwat.com/grpc-realtime-chat/pb/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
	defer conn.Close()

	client := pb.NewChatServiceClient(conn)
	stream, _ := client.ChatStream(context.Background())

	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF || err != nil {
				log.Println("Disconnected")
				return
			}
			log.Printf("[%s]: %s\n", msg.Sender, msg.Message)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	log.Print("Enter your name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	for {
		text, _ := reader.ReadString('\n')
		stream.Send(&pb.ChatMessage{Sender: name, Message: strings.TrimSpace(text)})
		time.Sleep(100 * time.Millisecond)
	}
}
