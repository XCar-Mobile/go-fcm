package main

import (
	"log"

	"github.com/XCar-Mobile/go-fcm"
)

func main() {
	// Create the message to be sent.
	msg := fcm.Message{
		Topic: "sample_device_token",
		Data: map[string]interface{}{
			"foo": "bar",
		},
		Notification: &fcm.Notification{
			Title: "title",
			Body:  "body",
		},
	}
	newMsg := &fcm.NewMessage{Message: msg}

	// Create a FCM client to send the message.
	client, err := fcm.NewClient("project_id")
	if err != nil {
		log.Fatalln(err)
	}

	// Send the message and receive the response without retries.
	token := "oauth_token"
	response, err := client.Send(newMsg, token)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%#v\n", response)
}
