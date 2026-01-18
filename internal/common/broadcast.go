package common

type BroadcastMessage struct {
	Room    string
	Message []byte
	Sender  *Client
}
