package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/onsi/gomega/types"
)

// HaveReceived returns a custom matcher that checks that the client has received at least one
// message of a given type.
//
// messageRef is a pointer to a message struct of the expected type. The matcher will set the value
// of the pointer to most recently received message of the type.
func HaveReceived(messageRef interface{}) types.GomegaMatcher {
	return &haveReceivedMatcher{messageRef: messageRef}
}

type haveReceivedMatcher struct {
	messageRef interface{}
	messages   []interface{}
}

func (h *haveReceivedMatcher) Match(actual interface{}) (success bool, err error) {
	client, ok := actual.(*Client)
	if !ok {
		return false, errors.New("haveReceivedMatcher expects a *Client")
	}

	if reflect.TypeOf(h.messageRef).Kind() != reflect.Ptr {
		return false, errors.New("haveReceivedMatcher messageRef must be a pointer")
	}

	h.messages = client.messagesSinceLastSend

	// Iterate in reverse so that we save the most recent matching message.
	for i := len(h.messages) - 1; i >= 0; i-- {
		msg := h.messages[i]
		if reflect.TypeOf(msg).Elem().AssignableTo(reflect.ValueOf(h.messageRef).Elem().Type()) {
			reflect.ValueOf(h.messageRef).Elem().Set(reflect.ValueOf(msg).Elem())
			return true, nil
		}
	}

	return false, nil
}

func (h *haveReceivedMatcher) FailureMessage(_ interface{}) (message string) {
	var trailer string

	if len(h.messages) == 0 {
		trailer = "0 messages received."
	} else {
		lastMessage := h.messages[len(h.messages)-1]

		lastMessageBytes, _ := json.Marshal(lastMessage)
		if lastMessageBytes == nil {
			lastMessageBytes = []byte("<could not parse message body>")
		}

		if len(h.messages) == 1 {
			trailer = fmt.Sprintf("1 message received. It had type %T: %s.", lastMessage, string(lastMessageBytes))
		} else {
			trailer = fmt.Sprintf("%d messages received. The last message had type %T: %s.", len(h.messages), lastMessage, string(lastMessageBytes))
		}
	}

	return fmt.Sprintf("No message was received with type %T. (%s)", h.messageRef, trailer)
}

func (h *haveReceivedMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	messageBytes, _ := json.Marshal(h.messageRef)
	if messageBytes == nil {
		messageBytes = []byte("<could not parse message body>")
	}

	return fmt.Sprintf("A message was received with type %T: %s.", h.messageRef, string(messageBytes))
}
