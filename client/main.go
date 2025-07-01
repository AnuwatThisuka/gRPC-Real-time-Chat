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

const (
	gap        = "\n\n"
	serverAddr = "localhost:50051"
	charLimit  = 280
	width      = 30
	height     = 3
	vpHeight   = 5
)

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	username    string
	client      pb.ChatServiceClient
	stream      pb.ChatService_ChatStreamClient
	err         error
}

type errMsg error

type messageReceivedMsg struct {
	sender  string
	message string
}

func main() {
	username := getUserInput("Enter your name: ")

	client, stream, err := setupGRPCConnection()
	if err != nil {
		log.Fatalf("Failed to setup gRPC connection: %v", err)
	}

	m := initialModel(username, client, stream)
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Start listening for messages
	go m.listenForMessages(p)

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}

func getUserInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func setupGRPCConnection() (pb.ChatServiceClient, pb.ChatService_ChatStreamClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial server: %w", err)
	}

	client := pb.NewChatServiceClient(conn)
	stream, err := client.ChatStream(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stream: %w", err)
	}

	return client, stream, nil
}

func initialModel(username string, client pb.ChatServiceClient, stream pb.ChatService_ChatStreamClient) model {
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

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		username:    username,
		client:      client,
		stream:      stream,
		err:         nil,
	}
}

func (m model) listenForMessages(p *tea.Program) {
	for {
		msg, err := m.stream.Recv()
		if err == io.EOF {
			log.Println("Server closed connection")
			return
		}
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			p.Send(errMsg(err))
			return
		}

		p.Send(messageReceivedMsg{
			sender:  msg.Sender,
			message: msg.Message,
		})
	}
}

func (m model) sendMessage(message string) error {
	if strings.TrimSpace(message) == "" {
		return nil
	}

	return m.stream.Send(&pb.ChatMessage{
		Sender:  m.username,
		Message: message,
	})
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)
		m.updateViewportContent()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if err := m.sendMessage(m.textarea.Value()); err != nil {
				m.err = err
			} else {
				m.addMessage("You", m.textarea.Value())
				m.textarea.Reset()
			}
		}

	case messageReceivedMsg:
		m.addMessage(msg.sender, msg.message)

	case errMsg:
		m.err = msg
		return m, tea.Quit
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *model) addMessage(sender, message string) {
	formattedMsg := m.senderStyle.Render(sender+": ") + message
	m.messages = append(m.messages, formattedMsg)
	m.updateViewportContent()
}

func (m *model) updateViewportContent() {
	if len(m.messages) > 0 {
		content := strings.Join(m.messages, "\n")
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(content))
	}
	m.viewport.GotoBottom()
}

func (m model) View() string {
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
