package config

// Config struct def
type Config struct {
	Database    string
	ThreadColl  string
	MessageColl string
}

// New returns a new config
func New() Config {
	return Config{
		Database:    "messaging",
		ThreadColl:  "thread",
		MessageColl: "message",
	}
}
