package model

import (
	"strconv"
	"time"

	"github.com/mitchellh/hashstructure"
)

// Thread entity definition
type Thread struct {
	ThreadID  string     `json:"threadId" bson:"threadId"`
	UserID1   string     `json:"userId1" bson:"userId1"`
	UserID2   string     `json:"userId2" bson:"userId2"`
	Messages  []*Message `json:"messages,omitempty" bson:"messages"`
	UpdatedAt time.Time  `json:"updatedAt" bson:"updatedAt"`
}

// GenerateThreadIDHash generates a thread id using two user ids
func GenerateThreadIDHash(u1, u2 string) (string, error) {
	type ComplexStruct struct {
		Data []string `hash:"set"`
	}

	v := ComplexStruct{
		Data: []string{
			u1, u2,
		},
	}

	hash, err := hashstructure.Hash(v, nil)
	if err != nil {
		panic(err)
	}

	h := strconv.FormatUint(hash, 10)
	return h, nil
}

// GetMessagesInThreadRequest defines entity for getting all messages in a thread
type GetMessagesInThreadRequest struct {
	ThreadID string `json:"threadId"`
	Limit    int    `json:"limit"`
	Skip     int    `json:"skip"`
}
