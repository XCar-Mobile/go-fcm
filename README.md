# go-fcm

[![GoDoc](https://godoc.org/github.com/appleboy/go-fcm?status.svg)](https://godoc.org/github.com/appleboy/go-fcm)
[![Build Status](https://cloud.drone.io/api/badges/appleboy/go-fcm/status.svg)](https://cloud.drone.io/appleboy/go-fcm)
[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/go-fcm)](https://goreportcard.com/report/github.com/appleboy/go-fcmm)

This project was forked from [github.com/edganiukov/fcm](https://github.com/edganiukov/fcm).

Golang client library for Firebase Cloud Messaging. Implemented only [HTTP client](https://firebase.google.com/docs/cloud-messaging/http-server-ref#downstream).

More information on [Firebase Cloud Messaging](https://firebase.google.com/docs/cloud-messaging/)

## Feature

* [x] Send messages to a topic
* [x] Send messages to a device list
* [x] Supports condition attribute (fcm only)

## Getting Started

To install fcm, use `go get`:

```bash
go get github.com/XCar-Mobile/go-fcm
```

or `govendor`:

```bash
govendor fetch github.com/XCar-Mobile/go-fcm
```

or other tool for vendoring.

## Sample Usage

Here is a simple example illustrating how to use FCM library:

```go
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
```
