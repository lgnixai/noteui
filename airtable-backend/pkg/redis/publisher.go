package redis

import (
	"log"
)

func Publish(channel string, message string) {
	err := RDB.Publish(Ctx, channel, message).Err()
	if err != nil {
		log.Printf("Error publishing message to channel %s: %v", channel, err)
	} else {
		log.Printf("Published message to channel %s", channel)
	}
}
