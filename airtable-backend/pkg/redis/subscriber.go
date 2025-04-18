package redis

import (
	"context" // Import context
	"log"

	"github.com/go-redis/redis/v8"
)

type Subscriber struct {
	*redis.PubSub
	// Add context here if needed, though we use the global Ctx
}

// NewSubscriber creates a new Redis subscriber instance.
func NewSubscriber() *Subscriber {
	// Note: RDB must be connected before calling this.
	// Subscribe initially to a dummy channel or none, actual subscriptions managed later.
	// Let's just initialize the PubSub object without subscribing yet.
	// The manager will handle the first subscriptions.
	return &Subscriber{
		PubSub: RDB.Subscribe(Ctx), // This returns the PubSub client for SUBSCRIBE
	}
}

// Subscribe adds channels to the subscriber.
func (s *Subscriber) Subscribe(channels ...string) error {
	// Corrected: Subscribe returns error directly
	log.Printf("Attempting Redis SUBSCRIBE to: %v", channels)
	err := s.PubSub.Subscribe(Ctx, channels...) // Subscribe returns *PubSub, err
	if err != nil {
		log.Printf("Error during Redis SUBSCRIBE call: %v", err)
		return err // Return the error from the method call
	}
	// Note: Successful return here doesn't mean the SUBSCRIBE command completed on Redis,
	// only that it was sent successfully over the connection.
	// Confirmation comes via the Receive methods receiving "subscribe" messages.
	log.Printf("Redis SUBSCRIBE command sent for: %v", channels)
	return nil // Return nil if the command was sent without immediate error
}

// Unsubscribe removes channels from the subscriber.
func (s *Subscriber) Unsubscribe(channels ...string) error {
	// Corrected: Unsubscribe returns error directly
	log.Printf("Attempting Redis UNSUBSCRIBE from: %v", channels)
	err := s.PubSub.Unsubscribe(Ctx, channels...) // Unsubscribe returns error
	if err != nil {
		log.Printf("Error during Redis UNSUBSCRIBE call: %v", err)
		return err // Return the error from the method call
	}
	// Confirmation comes via the Receive methods receiving "unsubscribe" messages.
	log.Printf("Redis UNSUBSCRIBE command sent for: %v", channels)
	return nil // Return nil if the command was sent without immediate error
}

// Listen starts listening for messages on the subscribed channels.
// It blocks until the context is cancelled or an error occurs.
// The messageHandler function is called for each received message.
func (s *Subscriber) Listen(messageHandler func(channel string, message string)) {
	// This method blocks, intended to be run in a goroutine.
	// The first successful Receive() or ReceiveMessage() after Subscribe
	// will confirm the subscription is active.
	log.Println("Redis subscriber starting message listener")

	for {
		// ReceiveMessage blocks until a message is available or an error occurs.
		msg, err := s.PubSub.ReceiveMessage(Ctx)
		if err != nil {
			// Check if the error is due to context cancellation (e.g., program shutdown)
			if err == context.Canceled {
				log.Println("Redis subscriber context cancelled, stopping listener")
				return // Exit the goroutine gracefully
			}
			log.Printf("Error receiving message from Redis Pub/Sub: %v", err)
			// Handle other errors (e.g., connection lost). Maybe attempt to reconnect
			// the PubSub client or log and continue looping depending on desired resilience.
			// For this example, we'll just log and continue, hoping the connection recovers.
			continue
		}

		// log.Printf("Received message from channel %s: %s", msg.Channel, msg.Payload) // Log inside the handler is better
		if messageHandler != nil {
			messageHandler(msg.Channel, msg.Payload)
		}
	}
}

// Close closes the subscriber connection.
func (s *Subscriber) Close() error {
	log.Println("Closing Redis subscriber connection")
	return s.PubSub.Close()
}
