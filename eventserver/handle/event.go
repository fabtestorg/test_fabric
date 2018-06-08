package handle

import "github.com/peersafe/factoring/apiserver/define"

type Event struct {
	Header   Header   `json:"header"`
	Contents Contents `json:"contents"`
}

type Header struct {
	ContentDef     ContentDef            `json:"contentDef"`
	Ack            Ack                   `json:"ack"`
	ResponseStatus define.ResponseStatus `json:"responseStatus"`
}

type ContentDef struct {
	ContentType string `json:"contentType"`
	TrackId     string `json:"trackId"`
	Language    string `json:"language"`
}

type Ack struct {
	Level    string `json:"level"`
	Callback string `json:"callback"`
}

type Contents struct {
	Schema  string      `json:"$schema"`
	Payload interface{} `json:"payload"`
	Command Command     `json:"command, omitempty"`
}

type Command struct {
	Uri    string `json:"uri, omitempty"`
	Action string `json:"action, omitempty"`
	Desc   string `json:"desc, omitempty"`
}

type ResponseData struct {
	TrackId        string                `json:"trackId"`
	ResponseStatus define.ResponseStatus `json:"responseStatus"`
	Page           define.Page           `json:"page"`
	Payload        interface{}           `json:"payload"`
}

type BlockInfo struct {
	Block_number uint64 `json:"block_number"`
	Tx_index     int    `json:"tx_index"`
}

type BlockInfoAll struct {
	BlockInfo
	MsgInfo        interface{} `json:"msgInfo"`
}