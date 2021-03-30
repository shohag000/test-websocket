package repository

import "github.com/shohag000/test-websocket/model"

// MessagingRepository defines the messaging repository
type MessagingRepository interface {
	GetInboxByUserID(userID string, messageLimit int64) (*model.Inbox, error)
	StoreMessage(message *model.Message) error
	StoreThread(thread *model.Thread) error
	FindThreadByUsers(userID, otherUserID string) (*model.Thread, error)
	GetAllThreadsByUserID(userID string) ([]*model.Thread, error)
	GetAllMessagesByThreadID(threadID string, limit, skip int64) ([]*model.Message, error)
}
