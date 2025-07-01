# gRPC Real-time Chat 💬

A real-time chat application built with Go and gRPC, featuring bidirectional streaming for instant messaging between multiple clients.

## 🚀 Features

- **Real-time messaging** using gRPC bidirectional streaming
- **Multi-client support** with concurrent user management
- **Simple CLI interface** for easy interaction
- **Protocol Buffers** for efficient message serialization
- **Docker support** for easy deployment
- **Thread-safe** client management with mutex locks

## 📋 Prerequisites

- Go 1.21 or higher
- Protocol Buffers compiler (`protoc`)
- Docker (optional, for containerized deployment)

### Installing protoc

**macOS:**

```bash
brew install protobuf
```

**Ubuntu/Debian:**

```bash
sudo apt update
sudo apt install protobuf-compiler
```

**Windows:**
Download from [Protocol Buffers releases](https://github.com/protocolbuffers/protobuf/releases)

## 🛠️ Installation

1. **Clone the repository:**

   ```bash
   git clone <repository-url>
   cd gRPC\ Real-time\ Chat
   ```

2. **Install Go dependencies:**

   ```bash
   go mod tidy
   ```

3. **Install protoc plugins:**

   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

4. **Generate Protocol Buffer code:**
   ```bash
   make generate
   ```

## 🏃‍♂️ Running the Application

### Method 1: Direct Go execution

1. **Start the server:**

   ```bash
   go run server/main.go
   ```

   The server will start on port `50051`

2. **Start clients** (in separate terminals):
   ```bash
   go run client/main.go
   ```

### Method 2: Using Docker

1. **Build and run the server:**

   ```bash
   docker build -t grpc-chat-server .
   docker run -p 50051:50051 grpc-chat-server
   ```

2. **Run clients locally:**
   ```bash
   go run client/main.go
   ```

## 📖 Usage

1. When you start a client, you'll be prompted to enter your name
2. Type messages and press Enter to send them to all connected users
3. Messages from other users will appear in real-time
4. Use `Ctrl+C` to disconnect

### Example Session

```
$ go run client/main.go
Enter your name: Alice
Hello everyone!
[Bob]: Hi Alice! 👋
[Charlie]: Welcome to the chat!
How's everyone doing?
[Bob]: Great! Thanks for asking
```

## 🏗️ Project Structure

```
.
├── Dockerfile              # Docker configuration for server
├── Makefile               # Build automation
├── README.md              # Project documentation
├── go.mod                 # Go module dependencies
├── go.sum                 # Go module checksums
├── client/
│   └── main.go           # Chat client implementation
├── server/
│   └── main.go           # Chat server implementation
├── proto/
│   └── chat.proto        # Protocol Buffer definitions
└── pb/
    └── proto/
        ├── chat.pb.go         # Generated protobuf code
        └── chat_grpc.pb.go    # Generated gRPC code
```

## 🔧 API Reference

### Protocol Buffer Schema

```protobuf
message ChatMessage {
  string sender = 1;    // Username of the message sender
  string message = 2;   // Message content
}

service ChatService {
  // Bidirectional streaming RPC for real-time chat
  rpc ChatStream (stream ChatMessage) returns (stream ChatMessage);
}
```

### Server Configuration

- **Port:** 50051
- **Protocol:** gRPC over TCP
- **Streaming:** Bidirectional

## 🐳 Docker Deployment

### Build the image:

```bash
docker build -t grpc-chat-server .
```

### Run the container:

```bash
docker run -p 50051:50051 grpc-chat-server
```

### Using Docker Compose (optional):

Create a `docker-compose.yml`:

```yaml
version: "3.8"
services:
  chat-server:
    build: .
    ports:
      - "50051:50051"
```

Then run:

```bash
docker-compose up
```

## 🧪 Development

### Regenerating Protocol Buffers

If you modify `proto/chat.proto`, regenerate the Go code:

```bash
make generate
```

### Adding New Features

1. Update the `.proto` file for new message types or services
2. Regenerate protobuf code with `make generate`
3. Implement the new functionality in server and client
4. Test with multiple clients

## 🚨 Troubleshooting

### Common Issues

**"command not found: protoc"**

- Install Protocol Buffers compiler (see Prerequisites)

**"plugin protoc-gen-go not found"**

- Install the protoc plugins:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```

**"connection refused"**

- Ensure the server is running on port 50051
- Check if another process is using the port:
  ```bash
  lsof -i :50051
  ```

**Client disconnects immediately**

- Check server logs for errors
- Ensure gRPC version compatibility

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and test thoroughly
4. Commit your changes: `git commit -am 'Add feature'`
5. Push to the branch: `git push origin feature-name`
6. Submit a pull request

## 📝 License

This project is open source and available under the [MIT License](LICENSE).

## 🔗 Resources

- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers Guide](https://developers.google.com/protocol-buffers)
- [Go gRPC Tutorial](https://grpc.io/docs/languages/go/quickstart/)

---

Made with ❤️ using Go and gRPC
