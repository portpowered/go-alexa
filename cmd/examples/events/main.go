// Package main demonstrates receiving Alexa endpoint events.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

const exampleSubscriptionDurationMinutes = 4
const errMissingBearerToken staticError = "Set ALEXA_BEARER_TOKEN"

type staticError string

func (err staticError) Error() string { return string(err) }

func main() {
	runErr := run()
	if runErr != nil {
		log.Fatal(runErr)
	}
}

func run() error {
	bearerToken := os.Getenv("ALEXA_BEARER_TOKEN")
	if bearerToken == "" {
		return errMissingBearerToken
	}

	client, err := alexa.NewClient()
	if err != nil {
		return fmt.Errorf("create Alexa client: %w", err)
	}

	session, err := client.NewSession(alexa.WithBearerToken(bearerToken))
	if err != nil {
		return fmt.Errorf("create Alexa session: %w", err)
	}

	defer func() { _ = session.Close() }()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	_, err = session.Subscribe(ctx, alexaapimodels.SubscribeRequest{
		Entities: []alexaapimodels.SubscribeEntity{{
			EntityType: alexaapimodels.EntityTypeEndpoint,
		}},
		DurationInMinutes: exampleSubscriptionDurationMinutes,
	})
	if err != nil {
		return fmt.Errorf("subscribe to endpoint events: %w", err)
	}

	outputLine("Connecting to event stream...")

	conn, err := session.ConnectEvents(ctx)
	if err != nil {
		return fmt.Errorf("connect to event stream: %w", err)
	}

	go func() {
		<-ctx.Done()

		_ = conn.Close()
	}()

	outputLine("Listening for events; press Ctrl+C to exit")

	for {
		event, err := conn.Receive()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			log.Printf("Receive event: %v", err)

			return nil
		}

		outputf("Event %s/%s for endpoint %s: %+v\n", event.Namespace, event.Name, event.EndpointID, event.Payload)
	}
}

func outputLine(values ...any) {
	_, err := fmt.Fprintln(os.Stdout, values...)
	if err != nil {
		panic(err)
	}
}

func outputf(format string, values ...any) {
	_, err := fmt.Fprintf(os.Stdout, format, values...)
	if err != nil {
		panic(err)
	}
}
