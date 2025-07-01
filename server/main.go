package main

import (
	"io"
	"log"
	"net"
	"sync"

	pb "anuwat.com/grpc-realtime-chat/pb/proto"

	"google.golang.org/grpc"
)

type chatServer struct {
	pb.UnimplementedChatServiceServer
	clients map[string]pb.ChatService_ChatStreamServer
	mu      sync.Mutex
}

func newChatServer() *chatServer {
	return &chatServer{clients: make(map[string]pb.ChatService_ChatStreamServer)}
}

func (s *chatServer) ChatStream(stream pb.ChatService_ChatStreamServer) error {
	var clientID string
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("recv error:", err)
			break
		}

		s.mu.Lock()
		if _, ok := s.clients[msg.Sender]; !ok {
			s.clients[msg.Sender] = stream
			clientID = msg.Sender
			log.Printf("👤 %s joined\n", clientID)
		}
		for id, cl := range s.clients {
			if id != msg.Sender {
				cl.Send(msg)
			}
		}
		s.mu.Unlock()
	}

	s.mu.Lock()
	delete(s.clients, clientID)
	s.mu.Unlock()
	log.Printf("👤 %s left\n", clientID)
	return nil
}

func main() {
	lis, _ := net.Listen("tcp", ":50051")
	s := grpc.NewServer()
	pb.RegisterChatServiceServer(s, newChatServer())
	log.Println("🚀 Server running on :50051")
	s.Serve(lis)
}
