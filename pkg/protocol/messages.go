package protocol

type Envelope struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

const (
	TypeJoinQueue   = "join_queue"
	TypeHeartbeat   = "heartbeat"
	TypeError       = "error"
	CodeBadFrame    = 1
	CodeUnknownType = 2
)
