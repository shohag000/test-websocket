package model

import (
	"time"
)

// // MessageType defines the type of the message
// type MessageType int

// const (
// 	// Text message type
// 	Text MessageType = iota
// 	// SwapAgreementCard message type
// 	SwapAgreementCard
// )

// func (d MessageType) String() string {
// 	return [...]string{"Text", "SwapAgreementCard"}[d]
// }

// var msgToString = map[MessageType]string{
// 	Text:              "Text",
// 	SwapAgreementCard: "SwapAgreementCard",
// }

// var msgToID = map[string]MessageType{
// 	"Text":              Text,
// 	"SwapAgreementCard": SwapAgreementCard,
// }

// // MarshalJSON marshals the enum as a quoted json string
// func (d MessageType) MarshalJSON() ([]byte, error) {
// 	buffer := bytes.NewBufferString(`"`)
// 	buffer.WriteString(msgToString[d])
// 	buffer.WriteString(`"`)
// 	return buffer.Bytes(), nil
// }

// // UnmarshalJSON unmashals a quoted json string to the enum value
// func (d *MessageType) UnmarshalJSON(b []byte) error {
// 	var j string
// 	err := json.Unmarshal(b, &j)
// 	if err != nil {
// 		return err
// 	}
// 	// Note that if the string cannot be found then it will be set to the zero value, 'Created' in this case.
// 	*d = msgToID[j]
// 	return nil
// }

// Message entity definition
type Message struct {
	ThreadID    string      `json:"threadId" bson:"threadId"`
	SenderID    string      `json:"senderId" bson:"senderId"`
	ReceiverID  string      `json:"receiverId" bson:"receiverId"`
	MessageType string      `json:"messageType" bson:"messageType"`
	MessageBody interface{} `json:"messageBody" bson:"messageBody"`
	CreatedAt   time.Time   `json:"createdAt" bson:"createdAt"`
}
