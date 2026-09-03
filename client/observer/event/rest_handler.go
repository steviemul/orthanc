package event

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const EVENT_PATH = "/events"

const (
	initialBackoff = 500 * time.Millisecond
	maxBackoff     = 30 * time.Second
	maxAttempts    = 6
)

type RestHandler struct {
	host   string
	client *http.Client
}

func (h RestHandler) getEventsPath() string {
	return h.host + EVENT_PATH
}

func (h RestHandler) HandleEvent(e Event) {
	go h.postEvent(e)
}

func (h RestHandler) postEvent(e Event) {
	eventJson, err := e.Json()

	if err != nil {
		log.Printf("Error reading event json [%s]", err)
		return
	}

	backoff := initialBackoff

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := h.sendEvent(eventJson)

		if err == nil {
			return
		}

		log.Printf("Error sending event request [%s, attempt=%d/%d, %s]", h.getEventsPath(), attempt, maxAttempts, err)

		if attempt == maxAttempts {
			log.Printf("Giving up sending event after %d attempts [%s]", maxAttempts, h.getEventsPath())
			return
		}

		jitter := time.Duration(rand.Int63n(int64(backoff)/2 + 1))
		time.Sleep(backoff + jitter)

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// sendEvent performs a single POST attempt. A non-nil error indicates a
// retryable failure (connection error or 5xx response). 4xx responses are
// treated as terminal and logged but not retried.
func (h RestHandler) sendEvent(body []byte) error {
	response, err := h.client.Post(
		h.getEventsPath(),
		"application/json",
		bytes.NewBuffer(body))

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode >= 500 {
		return fmt.Errorf("server error [%s]", response.Status)
	}

	if response.StatusCode >= 400 {
		log.Printf("Non-retryable response [%s, %s]", h.getEventsPath(), response.Status)
		return nil
	}

	log.Printf("Events request returned response [response=%s]", response.Status)

	return nil
}

func NewRestHandler(host string) RestHandler {
	return RestHandler{
		host: host,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}
