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

func main() {
	bearerToken := os.Getenv("ALEXA_BEARER_TOKEN")
	if bearerToken == "" {
		log.Fatal("Set ALEXA_BEARER_TOKEN")
	}

	client, err := alexa.NewClient()
	if err != nil {
		log.Fatalf("Create Alexa client: %v", err)
	}
	session, err := client.NewSession(alexa.WithBearerToken(bearerToken))
	if err != nil {
		log.Fatalf("Create Alexa session: %v", err)
	}
	defer func() { _ = session.Close() }()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	_, err = session.Subscribe(ctx, alexaapimodels.SubscribeRequest{
		Entities: []alexaapimodels.SubscribeEntity{{
			EntityType: alexaapimodels.EntityTypeEndpoint,
		}},
		DurationInMinutes: 4,
	})
	if err != nil {
		log.Fatalf("Subscribe to endpoint events: %v", err)
	}

	fmt.Println("Connecting to event stream...")
	conn, err := session.ConnectEvents(ctx)
	if err != nil {
		log.Fatalf("Connect to event stream: %v", err)
	}
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()

	fmt.Println("Listening for events; press Ctrl+C to exit")
	for {
		event, err := conn.Receive()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("Receive event: %v", err)
			return
		}
		fmt.Printf("Event %s/%s for endpoint %s: %+v\n", event.Namespace, event.Name, event.EndpointID, event.Payload)
	}
}
