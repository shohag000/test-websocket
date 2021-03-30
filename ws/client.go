// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/shohag000/test-websocket/batman/auth"
	"github.com/shohag000/test-websocket/handler"
	"github.com/shohag000/test-websocket/model"
	"github.com/shohag000/test-websocket/repository"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan model.Data

	// Authenticated bool type defins if the channel is authenticated with a valid user token
	Authenticated bool

	// MessagingService
	MessagingService handler.MessagingService

	// UserID is the user id of the client connected
	UserID string
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Trim message
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// Unmarshal incoming message
		incomingDataStr := string(message)
		var iData model.Data

		if err = json.Unmarshal([]byte(incomingDataStr), &iData); err != nil {
			fmt.Printf("could not unmarshal message: %v", err)
		}

		// Parse message based on data type
		switch iData.DataType {
		case model.InitData:
			// Initialize the connection, prior to this point the client is connected to the websocket,
			// however, the client is yet to be authenticated, without authentication the client will
			// not receive any kind of messages from the server.

			// Parse incoming json data
			var authMsg model.Auth
			err = mapstructure.Decode(iData.Data, &authMsg)
			if err != nil {
				fmt.Printf("could not parse data: %v", err)
				// Return error message
				c.hub.broadcast <- model.Data{
					DataType: model.ErrorData,
					Data: model.Error{
						Code:    "InvalidData",
						Details: fmt.Sprintf("Could not parse json data: %v", err),
					},
				}
				continue
			}

			// Validate auth token
			userID, ok, err := c.MessagingService.AuthenticateToken(authMsg.Token)
			if err != nil || !ok || userID != authMsg.UserID {
				// Return error message
				c.hub.broadcast <- model.Data{
					DataType: model.ErrorData,
					Data: model.Error{
						Code:    "InvalidToken",
						Details: fmt.Sprintf("Could not validate token: %v", err),
					},
				}
				continue
			}

			// Register client as authenticated
			c.UserID = userID
			c.Authenticated = true

			// Find user's inbox
			inbox, err := c.MessagingService.GetInboxByUserID(authMsg.UserID, 30)
			if err != nil {
				// Return error message
				c.hub.broadcast <- model.Data{
					DataType: model.ErrorData,
					Data: model.Error{
						Code:    "Internal",
						Details: fmt.Sprintf("Could not fetch inbox: %v", err),
					},
				}
				continue
			}

			// Return with user's inbox
			c.hub.broadcast <- model.Data{
				DataType: model.InboxData,
				Data:     inbox,
				UserID:   authMsg.UserID,
			}
			continue

		case model.MessageData:
			// Received message from the client, process the message, store it in database and send
			// it to the users websocket channel

			// TODO:: Validate message data
			// TODO:: Store message in database

			// Parse message data
			var msg model.Message
			err = mapstructure.Decode(iData.Data, &msg)
			if err != nil {
				fmt.Printf("could not parse msg data: %v", err)
				// Return error message
				c.hub.broadcast <- model.Data{
					DataType: model.ErrorData,
					Data: model.Error{
						Code:    "InvalidData",
						Details: fmt.Sprintf("Could not parse json data: %v", err),
					},
				}
				continue
			}

			// Set msg created at time
			msg.CreatedAt = time.Now()

			// Store message in database
			err = c.MessagingService.StoreMessage(&msg)
			if err != nil {
				// Return error message
				c.hub.broadcast <- model.Data{
					DataType: model.ErrorData,
					Data: model.Error{
						Code:    "Internal",
						Details: fmt.Sprintf("Could not save data: %v", err),
					},
				}
				continue
			}

			// Send data for broadcasting
			c.hub.broadcast <- model.Data{
				DataType: model.MessageData,
				Data:     msg,
				UserID:   msg.SenderID,
			}
			c.hub.broadcast <- model.Data{
				DataType: model.MessageData,
				Data:     msg,
				UserID:   msg.ReceiverID,
			}
			continue

		case model.ThreadData:
			// Parse message data
			var getAllMsgReq model.GetMessagesInThreadRequest
			err = mapstructure.Decode(iData.Data, &getAllMsgReq)
			if err != nil {
				fmt.Printf("could not parse GetMessagesInThreadRequest data: %v", err)
				// Return error message
				c.hub.broadcast <- model.Data{
					DataType: model.ErrorData,
					Data: model.Error{
						Code:    "InvalidData",
						Details: fmt.Sprintf("Could not parse json data: %v", err),
					},
				}
				continue
			}

			allMsg, err := c.MessagingService.GetAllMessagesByThreadID(
				getAllMsgReq.ThreadID,
				int64(getAllMsgReq.Limit),
				int64(getAllMsgReq.Skip),
			)
			if err != nil {
				// Return error message
				c.hub.broadcast <- model.Data{
					DataType: model.ErrorData,
					Data: model.Error{
						Code:    "Internal",
						Details: fmt.Sprintf("Could not fetch messages in thread: %v", err),
					},
				}
				continue
			}

			c.hub.broadcast <- model.Data{
				DataType: model.ThreadData,
				Data:     allMsg,
				UserID:   c.UserID,
			}
			continue

		default:
			// Handle invalid data type
			c.hub.broadcast <- model.Data{
				DataType: model.ErrorData,
				Data: model.Error{
					Code:    "InvalidDataType",
					Details: fmt.Sprintf("Invalid data type '%v' passed.", iData.DataType),
				},
			}
			continue
		}

		// Broadcast
		// c.hub.broadcast <- messageObj
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// Jsonify message data
			messageByte, err := json.Marshal(message)
			if err != nil {
				fmt.Printf("could not marshal data: %v", err)
			}
			w.Write(messageByte)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				// Jsonify message data
				messageByte, err := json.Marshal(<-c.send)
				if err != nil {
					fmt.Printf("could not marshal data: %v", err)
				}
				w.Write(messageByte)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Get db client
	dbClient, err := repository.GetDBClient()
	if err != nil {
		log.Println(err)
		return
	}

	// Get repository
	repo := repository.NewMongoRepository(dbClient)

	client := &Client{
		hub:              hub,
		conn:             conn,
		send:             make(chan model.Data, 256),
		Authenticated:    false,
		MessagingService: handler.NewService(repo, auth.New()),
		UserID:           "-1",
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
