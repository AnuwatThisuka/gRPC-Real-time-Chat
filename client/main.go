package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	pb "anuwat.com/grpc-realtime-chat/pb/proto"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Constants for UI configuration
const (
	gap        = "\n\n"
	serverAddr = "localhost:50051"
	charLimit  = 280
	width      = 30
	height     = 3
	vpHeight   = 5
)

// ChatClient wraps the gRPC client and stream
type ChatClient struct {
	client pb.ChatServiceClient
	stream pb.ChatService_ChatStreamClient
	conn   *grpc.ClientConn
}

// NewChatClient creates a new chat client with gRPC connection
func NewChatClient(serverAddr string) (*ChatClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}

	client := pb.NewChatServiceClient(conn)
	stream, err := client.ChatStream(context.Background())
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	return &ChatClient{
		client: client,
		stream: stream,
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection
func (c *ChatClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendMessage sends a message through the gRPC stream
func (c *ChatClient) SendMessage(sender, message string) error {
	if strings.TrimSpace(message) == "" {
		return nil
	}

	return c.stream.Send(&pb.ChatMessage{
		Sender:  sender,
		Message: message,
	})
}

// ReceiveMessages listens for incoming messages and sends them to the program
func (c *ChatClient) ReceiveMessages(p *tea.Program) {
	for {
		msg, err := c.stream.Recv()
		if err == io.EOF {
			log.Println("Server closed connection")
			p.Send(ConnectionClosedMsg{})
			return
		}
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			p.Send(ErrorMsg{err})
			return
		}

		p.Send(MessageReceivedMsg{
			Sender:  msg.Sender,
			Message: msg.Message,
		})
	}
}

// Message types for the Bubble Tea program
type (
	ErrorMsg            struct{ error }
	MessageReceivedMsg  struct{ Sender, Message string }
	ConnectionClosedMsg struct{}
)

// Model represents the application state
type Model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	username    string
	chatClient  *ChatClient
	err         error
}

// NewModel creates a new model with the given username and chat client
func NewModel(username string, chatClient *ChatClient) Model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = charLimit
	ta.SetWidth(width)
	ta.SetHeight(height)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(width, vpHeight)
	vp.SetContent("Welcome to the chat room!\nType a message and press Enter to send.")

	return Model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		username:    username,
		chatClient:  chatClient,
		err:         nil,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleWindowResize(msg)

	case tea.KeyMsg:
		cmd := m.handleKeyPress(msg)
		if cmd != nil {
			return m, cmd
		}

	case MessageReceivedMsg:
		m.addMessage(msg.Sender, msg.Message)

	case ErrorMsg:
		m.err = msg.error
		return m, tea.Quit

	case ConnectionClosedMsg:
		m.err = fmt.Errorf("connection closed by server")
		return m, tea.Quit
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

// handleWindowResize handles window resize events
func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) {
	m.viewport.Width = msg.Width
	m.textarea.SetWidth(msg.Width)
	m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)
	m.updateViewportContent()
}

// handleKeyPress handles key press events
func (m *Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return tea.Quit
	case tea.KeyEnter:
		message := m.textarea.Value()
		if err := m.chatClient.SendMessage(m.username, message); err != nil {
			m.err = err
		} else {
			m.addMessage("You", message)
			m.textarea.Reset()
		}
	}
	return nil
}

// addMessage adds a new message to the chat
func (m *Model) addMessage(sender, message string) {
	formattedMsg := m.senderStyle.Render(sender+": ") + message
	m.messages = append(m.messages, formattedMsg)
	m.updateViewportContent()
}

// updateViewportContent updates the viewport with the latest messages
func (m *Model) updateViewportContent() {
	if len(m.messages) > 0 {
		content := strings.Join(m.messages, "\n")
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(content))
	}
	m.viewport.GotoBottom()
}

// View implements tea.Model
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress any key to exit.", m.err)
	}

	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}

// getUserInput prompts the user for input and returns the trimmed string
func getUserInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func main() {
	username := getUserInput("Enter your name: ")
	if username == "" {
		log.Fatal("Username cannot be empty")
	}

	chatClient, err := NewChatClient(serverAddr)
	if err != nil {
		log.Fatalf("Failed to create chat client: %v", err)
	}
	defer chatClient.Close()

	model := NewModel(username, chatClient)
	program := tea.NewProgram(model, tea.WithAltScreen())

	// Start listening for messages in a goroutine
	go chatClient.ReceiveMessages(program)

	if _, err := program.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
