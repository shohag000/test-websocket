package model

// Data entity definition
type Data struct {
	DataType DataType    `json:"dataType"`
	Data     interface{} `json:"data"`
	UserID   string      `json:"-"`
}

// Auth data is passed from the client when authenticating the client
type Auth struct {
	Token  string `json:"token"`
	UserID string `json:"userId"`
}

// Error data is defines any errors sent from server to client
type Error struct {
	Details string `json:"details"`
	Code    string `json:"code"`
}
