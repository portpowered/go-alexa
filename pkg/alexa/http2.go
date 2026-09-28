package alexa

// Package connection provides HTTP/2 bidirectional communication for receiving real-time events from Alexa.
// It manages connection lifecycle, reconnection, and event streaming.
import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	directivewire "github.com/portpowered/go-alexa/pkg/alexa/internal/wire"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
	"golang.org/x/net/http2"
)

// HTTP2Connection implements the alexa.Connection interface for HTTP/2 event streaming
type HTTP2Connection struct {
	authority      string
	bearerToken    string
	tokenGetter    func(ctx context.Context) (string, error)
	client         *http.Client
	transport      *http2.Transport
	ctx            context.Context
	cancel         context.CancelFunc
	requestCancel  context.CancelFunc
	stopCloseHook  func() bool
	stopParentHook func() bool
	wg             sync.WaitGroup
	mu             sync.RWMutex
	response       *http.Response
	responseReader *bufio.Reader
	messageChan    chan *alexaapimodels.Event
	errChan        chan error
	boundary       string
	closed         bool
	channelsClosed bool
	lastPing       time.Time
}

// HTTP2ConnectionOption is a function that configures an HTTP2Connection
type HTTP2ConnectionOption func(*HTTP2Connection)

// WithTokenGetter sets a function to retrieve tokens dynamically
func WithTokenGetter(getter func(ctx context.Context) (string, error)) HTTP2ConnectionOption {
	return func(c *HTTP2Connection) {
		c.tokenGetter = getter
	}
}

// WithAuthority sets the HTTP/2 authority (hostname) for the connection
func WithAuthority(authority string) HTTP2ConnectionOption {
	return func(c *HTTP2Connection) {
		c.authority = authority
	}
}

// http2Connection creates a new HTTP/2 connection for event streaming
func http2Connection(opts ...HTTP2ConnectionOption) *HTTP2Connection {
	ctx, cancel := context.WithCancel(context.Background())

	transport := &http2.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		AllowHTTP: false,
	}

	conn := &HTTP2Connection{
		authority:   alexaapimodels.Http2ConnectionUriNa, // Default to NA
		transport:   transport,
		ctx:         ctx,
		cancel:      cancel,
		messageChan: make(chan *alexaapimodels.Event, 100),
		errChan:     make(chan error, 10),
		client: &http.Client{
			Transport: transport,
			Timeout:   0, // No timeout for long-lived connections
		},
	}

	for _, opt := range opts {
		opt(conn)
	}

	return conn
}

// newSessionHTTP2Connection creates an account-owned event stream and uses an
// injected HTTP client when the caller configured one.
func newSessionHTTP2Connection(authority string, tokenGetter func(context.Context) (string, error), client *http.Client) *HTTP2Connection {
	connection := http2Connection(WithAuthority(authority), WithTokenGetter(tokenGetter))
	if client != nil {
		connection.client = client
		connection.transport = nil
	}
	return connection
}

// getToken retrieves the bearer token
func (c *HTTP2Connection) getToken(ctx context.Context) (string, error) {
	if c.bearerToken != "" {
		return c.bearerToken, nil
	}
	if c.tokenGetter != nil {
		token, err := c.tokenGetter(ctx)
		if err != nil {
			return "", alexaapimodels.NewTokenError("failed to get token from token getter", err)
		}
		return token, nil
	}
	return "", alexaapimodels.NewTokenError("no token available", nil)
}

// Connect establishes the HTTP/2 connection to the directives endpoint
func (c *HTTP2Connection) Connect(ctx context.Context) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return alexaapimodels.NewClosedError("cannot connect: connection is already closed")
	}
	requestCtx, requestCancel := context.WithCancel(ctx)
	stopCloseHook := context.AfterFunc(c.ctx, requestCancel)
	stopParentHook := context.AfterFunc(requestCtx, c.cancel)
	c.requestCancel = requestCancel
	c.stopCloseHook = stopCloseHook
	c.stopParentHook = stopParentHook
	c.mu.Unlock()

	token, err := c.getToken(requestCtx)
	if err != nil {
		return err
	}

	// Build the URL for the directives endpoint
	url := fmt.Sprintf("https://%s/v20160207/directives", c.authority)

	// Create a request for the HTTP/2 connection
	req, err := http.NewRequestWithContext(requestCtx, "GET", url, nil)
	if err != nil {
		return alexaapimodels.NewConnectionError("failed to create request", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "*/*")

	// Perform the request
	resp, err := c.client.Do(req)
	if err != nil {
		return alexaapimodels.NewNetworkError("failed to connect to server", err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return alexaapimodels.NewAuthenticationError(string(body), resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return alexaapimodels.NewHTTPError(resp, string(body))
	}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		_ = resp.Body.Close()
		return alexaapimodels.NewClosedError("connection closed while connecting")
	}
	c.response = resp
	c.responseReader = bufio.NewReader(resp.Body)
	c.wg.Add(2)
	go c.processMessages()
	go c.managePings(requestCtx, token)
	c.mu.Unlock()

	return nil
}

// processMessages processes incoming messages from the streaming response
func (c *HTTP2Connection) processMessages() {
	defer c.wg.Done()
	defer func() {
		c.mu.Lock()
		if c.response != nil {
			_ = c.response.Body.Close()
		}
		c.mu.Unlock()
	}()

	reauthRequired := "Unable to authenticate the request. Please provide a valid authorization token."

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			c.mu.RLock()
			reader := c.responseReader
			c.mu.RUnlock()

			if reader == nil {
				return
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// Connection closed by server
					select {
					case c.errChan <- alexaapimodels.NewConnectionError("connection closed by server", err):
					default:
					}
					return
				}
				select {
				case c.errChan <- alexaapimodels.NewNetworkError("failed to read message", err):
				default:
				}
				return
			}

			line = strings.TrimSpace(line)

			// Handle boundary detection
			if strings.HasPrefix(line, "------") {
				if c.boundary == "" {
					c.boundary = line
				}
				continue
			}

			// Handle reauth errors
			if strings.Contains(line, reauthRequired) {
				select {
				case c.errChan <- alexaapimodels.NewAuthenticationError(line, http.StatusUnauthorized):
				default:
				}
				return
			}

			// Skip content-type headers
			if strings.HasPrefix(line, "Content-Type:") {
				continue
			}

			// Skip empty lines and boundary lines
			if line == "" || (c.boundary != "" && strings.HasPrefix(line, c.boundary)) {
				continue
			}

			// Decode the implementation-derived directive model. The schema does
			// not claim to be a provider-verified contract.
			var directiveMessage directivewire.DirectiveMessage
			if err := json.Unmarshal([]byte(line), &directiveMessage); err != nil {
				// Not JSON, skip
				continue
			}

			parsedEvent, err := parseDirectiveMessage(&directiveMessage)
			if err != nil {
				continue
			}
			select {
			case c.messageChan <- parsedEvent:
			case <-c.ctx.Done():
				return
			}
		}
	}
}

// managePings sends periodic ping requests to keep the connection alive
func (c *HTTP2Connection) managePings(ctx context.Context, token string) {
	defer c.wg.Done()

	ticker := time.NewTicker(299 * time.Second)
	defer ticker.Stop()

	// Send initial ping
	if err := c.ping(ctx, token); err != nil {
		select {
		case c.errChan <- err:
		default:
		}
	}

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if err := c.ping(ctx, token); err != nil {
				select {
				case c.errChan <- err:
				default:
				}
				// If ping fails with 403, it might be an auth error
				if alexaapimodels.IsHTTPStatusCode(err, http.StatusForbidden) || alexaapimodels.IsAuthenticationError(err) {
					return
				}
			}
		}
	}
}

// ping sends a ping request to the server
func (c *HTTP2Connection) ping(ctx context.Context, token string) error {
	url := fmt.Sprintf("https://%s/ping", c.authority)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return alexaapimodels.NewPingError(0, "failed to create ping request", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		return alexaapimodels.NewPingError(0, "ping request failed", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	c.mu.Lock()
	c.lastPing = time.Now()
	c.mu.Unlock()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		body, _ := io.ReadAll(resp.Body)
		return alexaapimodels.NewAuthenticationError(fmt.Sprintf("ping authentication failed: %s", string(body)), resp.StatusCode)
	}

	// Response status is 204 NO CONTENT
	if resp.StatusCode != http.StatusNoContent && (resp.StatusCode >= 400) {
		body, _ := io.ReadAll(resp.Body)
		return alexaapimodels.NewPingError(resp.StatusCode, fmt.Sprintf("ping returned status %d: %s", resp.StatusCode, string(body)), nil)
	}

	return nil
}

// Receive receives a decomposed event from the connection
// This implements the alexa.Connection interface
func (c *HTTP2Connection) Receive() (*alexaapimodels.Event, error) {
	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()

	if closed {
		return nil, alexaapimodels.NewClosedError("cannot receive: connection is closed")
	}

	select {
	case msg, ok := <-c.messageChan:
		if !ok {
			return nil, alexaapimodels.NewClosedError("message channel closed")
		}
		return msg, nil
	case err, ok := <-c.errChan:
		if !ok {
			return nil, alexaapimodels.NewClosedError("error channel closed")
		}
		return nil, err
	case <-c.ctx.Done():
		return nil, alexaapimodels.NewClosedError("context cancelled")
	}
}

// Send sends a message through the connection
// This implements the alexa.Connection interface
// Note: The Python implementation doesn't show sending messages, but we implement the interface
func (c *HTTP2Connection) Send(message *alexamodels.Message) error {
	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()

	if closed {
		return alexaapimodels.NewClosedError("cannot send: connection is closed")
	}

	// TODO: Implement message sending if needed
	// For now, this is a placeholder as the Python implementation doesn't show sending
	return alexaapimodels.NewConnectionError("sending messages not yet implemented", nil)
}

// Close closes the connection
// This implements the alexa.Connection interface
func (c *HTTP2Connection) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	response := c.response
	requestCancel := c.requestCancel
	stopCloseHook := c.stopCloseHook
	stopParentHook := c.stopParentHook
	c.mu.Unlock()

	c.cancel()
	if requestCancel != nil {
		requestCancel()
	}
	if stopCloseHook != nil {
		stopCloseHook()
	}
	if stopParentHook != nil {
		stopParentHook()
	}
	if response != nil {
		_ = response.Body.Close()
	}
	c.wg.Wait()

	c.mu.Lock()
	if c.response != nil {
		_ = c.response.Body.Close()
		c.response = nil
	}

	// Close channels safely (only once)
	if !c.channelsClosed {
		close(c.messageChan)
		close(c.errChan)
		c.channelsClosed = true
	}

	c.mu.Unlock()

	return nil
}

// ConnectionState represents the state of an HTTP/2 connection
type ConnectionState int

const (
	// ConnectionStateDisconnected indicates the connection is not established
	ConnectionStateDisconnected ConnectionState = iota
	// ConnectionStateConnecting indicates the connection is being established
	ConnectionStateConnecting
	// ConnectionStateConnected indicates the connection is active
	ConnectionStateConnected
	// ConnectionStateReconnecting indicates the connection is reconnecting
	ConnectionStateReconnecting
)
