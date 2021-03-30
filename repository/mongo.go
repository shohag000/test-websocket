package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shohag000/test-websocket/batman/database"
	"github.com/shohag000/test-websocket/batman/errorcodes"
	"github.com/shohag000/test-websocket/config"
	"github.com/shohag000/test-websocket/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type messagingRepository struct {
	client      *mongo.Client
	config      config.Config
	mongoHelper database.MongoHelper
}

func (mr *messagingRepository) GetInboxByUserID(userID string, messageLimit int64) (*model.Inbox, error) {
	// Create empty inbox
	inbox := model.Inbox{}

	// Fetch threads
	threads, err := mr.GetAllThreadsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch threads: %v", err)
	}

	// Fetch messages for each threads
	for _, tr := range threads {

		msgs, err := mr.GetAllMessagesByThreadID(tr.ThreadID, messageLimit, 0)
		if err != nil {
			continue
		}

		tr.Messages = msgs
	}

	// Add threads to inbox
	inbox.Threads = threads

	return &inbox, nil
}

func (mr *messagingRepository) StoreMessage(message *model.Message) error {
	err := mr.mongoHelper.Store(mr.config.Database, mr.config.MessageColl, message)
	if err != nil {
		return err
	}

	return nil
}

func (mr *messagingRepository) StoreThread(thread *model.Thread) error {
	err := mr.mongoHelper.Store(mr.config.Database, mr.config.ThreadColl, thread)
	if err != nil {
		return err
	}

	return nil
}

func (mr *messagingRepository) FindThreadByUsers(uID1, uID2 string) (*model.Thread, error) {
	// Generate users hash
	tID, err := model.GenerateThreadIDHash(uID1, uID2)
	if err != nil {
		return nil, fmt.Errorf("could not generate thread id hash: %v", err)
	}

	// Fetch data from database
	result := mr.mongoHelper.Fetch(mr.config.Database, mr.config.ThreadColl, tID, "threadId")
	thread := model.Thread{}
	err = result.Decode(&thread)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errorcodes.ErrNotFound
		}
		return nil, err
	}

	return &thread, nil
}

func (mr *messagingRepository) GetAllThreadsByUserID(userID string) ([]*model.Thread, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	collection := mr.client.Database(mr.config.Database).Collection(mr.config.ThreadColl)
	filter := bson.M{
		"$or": []interface{}{
			bson.M{"userId1": userID},
			bson.M{"userId2": userID},
		},
	}

	var results []*model.Thread

	cur, err := collection.Find(ctx, filter, &options.FindOptions{
		// Limit: &limit,
		// Skip:  &skip,
		Sort: bson.D{
			primitive.E{Key: "updatedAt", Value: -1},
		},
	})
	if err != nil {
		return results, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var elem model.Thread
		err := cur.Decode(&elem)
		if err != nil {
			continue
		}
		results = append(results, &elem)
	}

	return results, nil
}

func (mr *messagingRepository) GetAllMessagesByThreadID(threadID string, limit, skip int64) ([]*model.Message, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	collection := mr.client.Database(mr.config.Database).Collection(mr.config.MessageColl)
	filter := bson.M{
		"threadId": threadID,
	}

	var results []*model.Message

	cur, err := collection.Find(ctx, filter, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort: bson.D{
			primitive.E{Key: "createdAt", Value: -1},
		},
	})
	if err != nil {
		return results, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var elem model.Message
		err := cur.Decode(&elem)
		if err != nil {
			continue
		}
		results = append(results, &elem)
	}

	return results, nil
}

// NewMongoRepository returns a new mongo messaging repository
func NewMongoRepository(dbClient *mongo.Client) MessagingRepository {
	mHelper := database.NewMongoHelper(dbClient)
	return &messagingRepository{
		client:      dbClient,
		config:      config.New(),
		mongoHelper: mHelper,
	}
}

// GetDBClient returns a mongo client
func GetDBClient() (*mongo.Client, error) {
	cs := "mongodb://mongo:27017"
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cs))
	if err != nil {
		return nil, err
	}
	// defer client.Disconnect(ctx)

	return client, nil
}
