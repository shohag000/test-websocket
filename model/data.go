package model

import (
	"bytes"
	"encoding/json"
)

// DataType defines the type of the message
type DataType int

const (
	// InboxData message type defines the whole inbox of a user
	InboxData DataType = iota
	// MessageData message type defines a single message sent/received by a user
	MessageData
	// InitData message types defines data for initialization and authentication
	InitData
	// ThreadData message type defines all data in a thread
	ThreadData
	// ErrorData message type defines any error message
	ErrorData
)

func (d DataType) String() string {
	return [...]string{"InboxData", "MessageData", "InitData", "ErrorData"}[d]
}

var toString = map[DataType]string{
	InboxData:   "InboxData",
	MessageData: "MessageData",
	InitData:    "InitData",
	ThreadData:  "ThreadData",
	ErrorData:   "ErrorData",
}

var toID = map[string]DataType{
	"InboxData":   InboxData,
	"MessageData": MessageData,
	"InitData":    InitData,
	"ThreadData":  ThreadData,
	"ErrorData":   ErrorData,
}

// MarshalJSON marshals the enum as a quoted json string
func (d DataType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toString[d])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (d *DataType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Created' in this case.
	*d = toID[j]
	return nil
}
