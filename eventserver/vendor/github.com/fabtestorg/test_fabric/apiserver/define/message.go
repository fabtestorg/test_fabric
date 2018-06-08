package define

// Message 消息转发平台 结构
type Message struct {
	Factor                    // 业务逻辑数据
	PeersafeData PeersafeData `json:"peersafeData"`
}

type PeersafeData struct {
	Keys map[string][]byte `json:"keys"`
}

type QueryContents struct {
	Schema  string      `json:"$schema"`
	Payload interface{} `json:"payload"`
}
