package fcm

import "testing"

func TestValidate(t *testing.T) {
	t.Run("valid with token", func(t *testing.T) {
		timeToLive := uint(3600)
		msg := Message{
			Topic:      "test",
			TimeToLive: &timeToLive,
			Data: map[string]interface{}{
				"message": "This is a Firebase Cloud Messaging Topic Message!",
			},
		}
		newMsg := &NewMessage{Message: msg}
		err := newMsg.Validate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid message", func(t *testing.T) {
		var msg Message
		newMsg := &NewMessage{Message: msg}
		err := newMsg.Validate()
		if err == nil {
			t.Fatalf("expected <%v> error, but got <nil>", ErrInvalidMessage)
		}
	})

	t.Run("invalid target", func(t *testing.T) {
		msg := Message{
			Data: map[string]interface{}{
				"message": "This is a Firebase Cloud Messaging Topic Message!",
			},
		}
		newMsg := &NewMessage{Message: msg}
		err := newMsg.Validate()
		if err == nil {
			t.Fatalf("expected <%v> error, but got nil", ErrInvalidTarget)
		}
	})

	t.Run("valid with condition", func(t *testing.T) {
		msg := Message{
			Condition: "'dogs' in topics || 'cats' in topics",
			Data: map[string]interface{}{
				"message": "This is a Firebase Cloud Messaging Topic Message!",
			},
		}
		newMsg := &NewMessage{Message: msg}
		err := newMsg.Validate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid condition", func(t *testing.T) {
		msg := Message{
			Condition: "'TopicA' in topics && 'TopicB' in topics && 'TopicC' in topics && 'TopicD' in topics && 'TopicE' in topics && 'TopicF' in topics && 'TopicG' in topics && 'TopicH' in topics",
			Data: map[string]interface{}{
				"message": "This is a Firebase Cloud Messaging Topic Message!",
			},
		}
		newMsg := &NewMessage{Message: msg}
		err := newMsg.Validate()
		if err == nil {
			t.Fatalf("expected <%v> error, but got nil", ErrInvalidTarget)
		}
	})
}
