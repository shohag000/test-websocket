package handler

import (
	"errors"
	"fmt"
	"time"

	"github.com/shohag000/test-websocket/batman/auth"
	"github.com/shohag000/test-websocket/batman/errorcodes"
	"github.com/shohag000/test-websocket/model"
	"github.com/shohag000/test-websocket/repository"
)

// MessagingService defines the services of the messagins system
type MessagingService interface {
	AuthenticateToken(token string) (userID string, valid bool, err error)
	GetInboxByUserID(userID string, messageLimit int) (*model.Inbox, error)
	StoreMessage(message *model.Message) error
	// CreateThread(thread *model.Thread) error
	FindThreadByUsers(userID, otherUserID string) (*model.Thread, error)
	// FindThreadByThreadID(threadID string) (*model.Thread, error)
	GetAllMessagesByThreadID(threadID string, limit, skip int64) ([]*model.Message, error)
}

type messagingService struct {
	repo          repository.MessagingRepository
	authenticator auth.Authenticator
}

func (ms *messagingService) AuthenticateToken(token string) (string, bool, error) {
	if token == "a51f5e8cefa63a4a8ee4e2e34329661c" { //TODO:: remove this token
		return "system", true, nil
	}
	u, err := ms.authenticator.DecodeToken(token)
	if err != nil {
		return "", false, err
	}
	return u.UserID, true, nil
}

func (ms *messagingService) GetInboxByUserID(userID string, messageLimit int) (*model.Inbox, error) {
	inbox, err := ms.repo.GetInboxByUserID(userID, int64(messageLimit))
	if err != nil {
		return nil, fmt.Errorf("could not fetch inbox: %v", err)
	}
	return inbox, nil
}

func (ms *messagingService) StoreMessage(message *model.Message) error {
	// Check if thread exists, if not, create a new thread
	thread, err := ms.FindThreadByUsers(message.SenderID, message.ReceiverID)
	if err != nil {
		if !errors.Is(err, errorcodes.ErrNotFound) {
			return fmt.Errorf("could not find thread: %v", err)
		}
	}

	if thread == nil {
		// Generate new thread id
		tID, err := model.GenerateThreadIDHash(message.SenderID, message.ReceiverID)
		if err != nil {
			return fmt.Errorf("could not create thread id: %v", err)
		}

		// Create thread model
		thread = &model.Thread{
			ThreadID:  tID,
			UserID1:   message.SenderID,
			UserID2:   message.ReceiverID,
			UpdatedAt: time.Now(),
		}

		// Store thread
		err = ms.repo.StoreThread(thread)
		if err != nil {
			return fmt.Errorf("could not store thread: %v", err)
		}
	}

	// Store message
	message.ThreadID = thread.ThreadID
	err = ms.repo.StoreMessage(message)
	if err != nil {
		return fmt.Errorf("could not store message: %v", err)
	}

	return nil
}

// func (ms *messagingService) CreateThread(thread *model.Thread) error {
// 	return nil
// }

func (ms *messagingService) FindThreadByUsers(uID1, uID2 string) (*model.Thread, error) {
	tr, err := ms.repo.FindThreadByUsers(uID1, uID2)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

func (ms *messagingService) GetAllMessagesByThreadID(threadID string, limit, skip int64) ([]*model.Message, error) {
	messages, err := ms.repo.GetAllMessagesByThreadID(threadID, limit, skip)
	if err != nil {
		return nil, fmt.Errorf("could not fetch messages: %v", err)
	}

	return messages, nil
}

// NewService  returns a new messaging service
func NewService(repo repository.MessagingRepository, authenticator auth.Authenticator) MessagingService {
	return &messagingService{
		repo:          repo,
		authenticator: authenticator,
	}
}
