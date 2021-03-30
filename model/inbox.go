package model

// Inbox entity definitation
type Inbox struct {
	Threads []*Thread `json:"threads,omitempty" bson:"threads"`
}
