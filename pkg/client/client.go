package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/gorilla/websocket"
)

type Event struct {
	ID     string          `json:"id"`
	Time   time.Time       `json:"time"`
	Source string          `json:"source"`
	Type   string          `json:"type"`
	Device EventDeviceData `json:"data"`
}

type EventDeviceData struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	DetailedType string                 `json:"deviceType"`
	IsReachable  bool                   `json:"isReachable"`
	LastSeen     time.Time              `json:"lastSeen"`
	Attributes   map[string]interface{} `json:"attributes"`
}

type EventHandler func(Event)

type handlerRegistration struct {
	Handler EventHandler
	Types   []string
}

type Device struct {
}

type Client interface {
	ListDevices() ([]*Device, error)
	RegisterEventHandler(handler EventHandler, eventTypes ...string)
	SetEventLog(writer io.Writer)
	ListenForEvents() error
	StopEventListening() error
	GetEventLoopState() error
}

type client struct {
	httpClient          *http.Client
	authorization       *Authorization
	endpoint            string
	registrations       []handlerRegistration
	eventLoopRunning    bool
	eventLoopError      error
	eventLog            io.Writer
	websocketDialer     *websocket.Dialer
	websocketConnection *websocket.Conn
}

// Connect creates a new Client and provides functions to communicate with the IKEA Smart-Home hub.
func Connect(address string, port int, authorization *Authorization) Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify:    true,
		VerifyPeerCertificate: fingerprintVerifier(authorization, false),
	}
	transport := &authorizationRoundTripper{
		authorization: authorization,
		origin: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	return &client{
		authorization: authorization,
		endpoint:      fmt.Sprintf("%s:%d/v1", address, port),
		httpClient: &http.Client{
			Transport: transport,
		},
		websocketDialer: &websocket.Dialer{
			TLSClientConfig: tlsConfig,
		},
		eventLoopRunning: false,
		eventLog:         os.Stdout,
	}
}

func (c *client) ListDevices() ([]*Device, error) {
	targetURL := fmt.Sprintf("https://%s/devices", c.endpoint)
	response, err := c.httpClient.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error listing devices from %s: %w", targetURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error listing devices from %s: Received status code %d", targetURL, response.StatusCode)
	}

	var devices []*Device
	if err := json.NewDecoder(response.Body).Decode(&devices); err != nil {
		return nil, fmt.Errorf("error decoding devices response: %w", err)
	}

	return devices, nil
}

func (c *client) RegisterEventHandler(handler EventHandler, eventTypes ...string) {
	c.registrations = append(c.registrations, handlerRegistration{
		Handler: handler,
		Types:   eventTypes,
	})
}

func (c *client) SetEventLog(writer io.Writer) {
	c.eventLog = writer
}

func (c *client) ListenForEvents() error {
	if c.eventLoopRunning {
		return fmt.Errorf("Event loop already running")
	}
	c.eventLoopRunning = true
	return retry.Do(c.eventLoop, retry.DelayType(func(n uint, loopErr error, config *retry.Config) time.Duration {
		c.eventLoopError = loopErr
		_, _ = fmt.Fprintf(c.eventLog, "Error in event loop: %v\nRestarting event loop in 30 seconds\n", loopErr)
		return 30 * time.Second
	}), retry.Attempts(0))
}

func (c *client) eventLoop() error {
	websocketURL := fmt.Sprintf("wss://%s", c.endpoint)
	websocketHeader := http.Header{}
	websocketHeader.Set("Authorization", "Bearer "+c.authorization.AccessToken)
	c.eventLoopError = nil // Reset error cache
	var err error
	c.websocketConnection, _, err = c.websocketDialer.Dial(websocketURL, websocketHeader)
	if err != nil {
		return err
	}
	defer func(conn *websocket.Conn) {
		_, _ = fmt.Fprintf(c.eventLog, "\U0001F6AB Closing connection to %s\n", conn.RemoteAddr().String())
		_ = conn.Close()
	}(c.websocketConnection)
	_, _ = fmt.Fprintf(c.eventLog, "\U0001F50C Established connection to %v\n", c.websocketConnection.RemoteAddr())
	for {
		event := &Event{}
		err := c.websocketConnection.ReadJSON(event)
		if err != nil {
			if c.eventLoopRunning {
				return err
			} else {
				return nil // Error occurred because of terminating the event loop - returning without error
			}
		}
		for _, registration := range c.registrations {
			if len(registration.Types) == 0 || slices.Contains(registration.Types, event.Type) {
				registration.Handler(*event)
			}
		}
	}
}

func (c *client) GetEventLoopState() error {
	return c.eventLoopError
}

func (c *client) StopEventListening() error {
	if c.eventLoopRunning {
		c.eventLoopRunning = false
		if c.websocketConnection != nil {
			return c.websocketConnection.Close()
		}
	}
	return nil
}
