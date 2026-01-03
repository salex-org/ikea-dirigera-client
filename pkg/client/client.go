package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Event struct {
	ID     string    `json:"id"`
	Time   time.Time `json:"time"`
	Source string    `json:"source"`
	Type   string    `json:"type"`
	Device Device    `json:"data"`
}

type Device struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	DetailedType string                 `json:"deviceType"`
	IsReachable  bool                   `json:"isReachable"`
	CreatedAt    time.Time              `json:"createdAt"`
	LastSeen     time.Time              `json:"lastSeen"`
	Attributes   map[string]interface{} `json:"attributes"`
	Room         Room                   `json:"room"`
}

type Room struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Scene struct {
	ID        string    `json:"id"`
	Info      Info      `json:"info"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
	Triggers  []Trigger `json:"triggers"`
	Actions   []Action  `json:"actions"`
}

type Trigger struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Disabled bool   `json:"disabled"`
}

type Action struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	DeviceID   string                 `json:"deviceId"`
	Attributes map[string]interface{} `json:"attributes"`
}

type Info struct {
	Name string `json:"name"`
}

type User struct {
	ID        string    `json:"uid"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdTimestamp"`
}

type EventHandler func(Event)

type handlerRegistration struct {
	Handler EventHandler
	Types   []string
}

type Client interface {
	ListDevices() ([]*Device, error)
	GetDevice(deviceID string) (*Device, error)
	GetHubStatus() (*Device, error)
	ListRooms() ([]*Room, error)
	GetRoom(roomID string) (*Room, error)
	ListScenes() ([]*Scene, error)
	GetScene(sceneID string) (*Scene, error)
	ListUsers() ([]*User, error)
	GetUser(userID string) (*User, error)
	GetCurrentUser() (*User, error)
	DeleteUser(userID string) error
	RegisterEventHandler(handler EventHandler, eventTypes ...string)
	SetEventLog(writer io.Writer)
	ListenForEvents() error
	StopEventListening() error
	GetEventLoopState() error
	Get(url string) (string, error)
}

type client struct {
	httpClient          *http.Client
	authorization       *Authorization
	endpoint            string
	registrations       []handlerRegistration
	eventLoopMutex      sync.Mutex
	eventLoopContext    context.Context
	eventLoopCancelFunc context.CancelFunc
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
		eventLoopContext:    nil,
		eventLoopCancelFunc: nil,
		eventLog:            os.Stdout,
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

func (c *client) GetHubStatus() (*Device, error) {
	targetURL := fmt.Sprintf("https://%s/hub/status", c.endpoint)
	response, err := c.httpClient.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error reading hub status from %s: %w", targetURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error reading hub status from %s: Received status code %d", targetURL, response.StatusCode)
	}

	var device *Device
	if err := json.NewDecoder(response.Body).Decode(&device); err != nil {
		return nil, fmt.Errorf("error decoding hub status response: %w", err)
	}

	return device, nil
}

func (c *client) GetDevice(deviceID string) (*Device, error) {
	targetURL := fmt.Sprintf("https://%s/devices/%s", c.endpoint, deviceID)
	response, err := c.httpClient.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error reading device %s from %s: %w", deviceID, targetURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error reading device %s from %s: Received status code %d", deviceID, targetURL, response.StatusCode)
	}

	var device *Device
	if err := json.NewDecoder(response.Body).Decode(&device); err != nil {
		return nil, fmt.Errorf("error decoding get device response: %w", err)
	}

	return device, nil
}

func (c *client) ListRooms() ([]*Room, error) {
	targetURL := fmt.Sprintf("https://%s/rooms", c.endpoint)
	response, err := c.httpClient.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error listing rooms from %s: %w", targetURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error listing rooms from %s: Received status code %d", targetURL, response.StatusCode)
	}

	var rooms []*Room
	if err := json.NewDecoder(response.Body).Decode(&rooms); err != nil {
		return nil, fmt.Errorf("error decoding rooms response: %w", err)
	}

	return rooms, nil
}

func (c *client) GetRoom(roomID string) (*Room, error) {
	targetURL := fmt.Sprintf("https://%s/rooms/%s", c.endpoint, roomID)
	response, err := c.httpClient.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error reading room %s from %s: %w", roomID, targetURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error reading room %s from %s: Received status code %d", roomID, targetURL, response.StatusCode)
	}

	var room *Room
	if err := json.NewDecoder(response.Body).Decode(&room); err != nil {
		return nil, fmt.Errorf("error decoding get room response: %w", err)
	}

	return room, nil
}

func (c *client) ListScenes() ([]*Scene, error) {
	targetURL := fmt.Sprintf("https://%s/scenes", c.endpoint)
	response, err := c.httpClient.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error listing scenes from %s: %w", targetURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error listing scenes from %s: Received status code %d", targetURL, response.StatusCode)
	}

	var scenes []*Scene
	if err := json.NewDecoder(response.Body).Decode(&scenes); err != nil {
		return nil, fmt.Errorf("error decoding scenes response: %w", err)
	}

	return scenes, nil
}

func (c *client) GetScene(sceneID string) (*Scene, error) {
	targetURL := fmt.Sprintf("https://%s/scenes/%s", c.endpoint, sceneID)
	response, err := c.httpClient.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error reading scene %s from %s: %w", sceneID, targetURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error reading scene %s from %s: Received status code %d", sceneID, targetURL, response.StatusCode)
	}

	var scene *Scene
	if err := json.NewDecoder(response.Body).Decode(&scene); err != nil {
		return nil, fmt.Errorf("error decoding get scene response: %w", err)
	}

	return scene, nil
}

func (c *client) ListUsers() ([]*User, error) {
	targetURL := fmt.Sprintf("https://%s/users", c.endpoint)
	response, err := c.httpClient.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error listing users from %s: %w", targetURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error listing users from %s: Received status code %d", targetURL, response.StatusCode)
	}

	var users []*User
	if err := json.NewDecoder(response.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("error decoding users response: %w", err)
	}

	return users, nil
}

func (c *client) GetUser(userID string) (*User, error) {
	targetURL := fmt.Sprintf("https://%s/users/%s", c.endpoint, userID)
	response, err := c.httpClient.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error reading user %s from %s: %w", userID, targetURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error reading user %s from %s: Received status code %d", userID, targetURL, response.StatusCode)
	}

	var user *User
	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error decoding get user response: %w", err)
	}

	return user, nil
}

func (c *client) GetCurrentUser() (*User, error) {
	targetURL := fmt.Sprintf("https://%s/users/me", c.endpoint)
	response, err := c.httpClient.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error reading current user from %s: %w", targetURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error reading current user from %s: Received status code %d", targetURL, response.StatusCode)
	}

	var user *User
	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error decoding get current user response: %w", err)
	}

	return user, nil
}

func (c *client) DeleteUser(userID string) error {
	targetURL := fmt.Sprintf("https://%s/users/%s", c.endpoint, userID)
	request, err := http.NewRequest("DELETE", targetURL, nil)
	if err != nil {
		return fmt.Errorf("error creating delete call for user %s: %w", userID, err)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("error removing user %s from %s: %w", userID, targetURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error removing user %s from %s: Received status code %d", userID, targetURL, response.StatusCode)
	}

	return nil
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
	c.eventLoopMutex.Lock()
	if c.eventLoopContext != nil {
		c.eventLoopMutex.Unlock()
		return fmt.Errorf("Event loop already running")
	}
	c.eventLoopContext, c.eventLoopCancelFunc = context.WithCancel(context.Background())
	c.eventLoopError = nil
	c.eventLoopMutex.Unlock()

	for {
		if err := c.eventLoop(); err != nil {
			c.eventLoopMutex.Lock()
			c.eventLoopError = err
			c.eventLoopMutex.Unlock()

			// "Echter" Fehler nur, wenn der EventLoopContext nich beendet wurde
			select {
			case <-c.eventLoopContext.Done():
			default:
				_, _ = fmt.Fprintf(c.eventLog, "Error in event loop: %v\nRestarting event loop in 30 seconds\n", err)
			}

			timer := time.NewTimer(30 * time.Second)
			select {
			case <-timer.C:

				c.eventLoopMutex.Lock()
				c.eventLoopError = nil
				c.eventLoopMutex.Unlock()

				continue
			case <-c.eventLoopContext.Done():
				if !timer.Stop() {
					<-timer.C
				}

				c.eventLoopMutex.Lock()
				c.eventLoopError = nil
				c.eventLoopCancelFunc = nil
				c.eventLoopContext = nil
				c.eventLoopMutex.Unlock()

				return nil
			}
		}
	}
}

func (c *client) eventLoop() error {
	var err error

	websocketURL := fmt.Sprintf("wss://%s", c.endpoint)
	websocketHeader := http.Header{}
	websocketHeader.Set("Authorization", "Bearer "+c.authorization.AccessToken)

	c.eventLoopMutex.Lock()
	c.websocketConnection, _, err = c.websocketDialer.Dial(websocketURL, websocketHeader)
	c.eventLoopMutex.Unlock()
	if err != nil {
		return err
	}
	defer func(conn *websocket.Conn) {
		_, _ = fmt.Fprintf(c.eventLog, "\U0001F6AB Closing connection to %s\n", conn.RemoteAddr().String())
		_ = conn.Close()

		c.eventLoopMutex.Lock()
		c.websocketConnection = nil
		c.eventLoopMutex.Unlock()
	}(c.websocketConnection)
	_, _ = fmt.Fprintf(c.eventLog, "\U0001F50C Established connection to %v\n", c.websocketConnection.RemoteAddr())
	for {
		event := &Event{}
		err := c.websocketConnection.ReadJSON(event)
		if err != nil {
			return err
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
	c.eventLoopMutex.Lock()
	defer c.eventLoopMutex.Unlock()

	if c.eventLoopCancelFunc != nil {
		c.eventLoopCancelFunc()
		if c.websocketConnection != nil {
			err := c.websocketConnection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				return err
			}

			return c.websocketConnection.Close()
		}
	}

	return nil
}

func (c *client) Get(path string) (string, error) {
	base, err := url.Parse(fmt.Sprintf("https://%s", c.endpoint))
	if err != nil {
		return "", fmt.Errorf("error parsing address %q: %w", c.endpoint, err)
	}
	targetURL := base.JoinPath(path).String()
	response, err := c.httpClient.Get(targetURL)
	if err != nil {
		return "", fmt.Errorf("error calling %s: %w", targetURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error calling %s: Received status code %d", targetURL, response.StatusCode)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	return string(bodyBytes), nil
}
