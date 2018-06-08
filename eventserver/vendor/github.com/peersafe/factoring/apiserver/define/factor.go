package define

type Factor struct {
	CreateBy        string   `json:"createBy"`        // 创建者
	CreateTime      uint64   `json:"createTime"`      // 创建时间
	Sender          string   `json:"sender"`          // 发送者
	Receiver        []string `json:"receiver"`        // 接收者列表
	TxData          string   `json:"txData"`          // 业务数据
	LastUpdateTime  uint64   `json:"lastUpdateTime"`  // 最近一次修改时间
	LastUpdateBy    string   `json:"lastUpdateBy"`    // 最近一次修改者
	CryptoFlag      int      `json:"cryptoFlag"`      // 加密标识（0:不加密，1:加密）
	CryptoAlgorithm string   `json:"cryptoAlgorithm"` // 加密算法类型
	DocType         string   `json:"docType"`         // 业务类型
	FabricTxId      string   `json:"fabricTxId"`      // Fabric交易id(uuid)
	BusinessNo      string   `json:"businessNo"`      // 业务编号（交易编号）
	Expand1         string   `json:"expand1"`         // 扩展字段1
	Expand2         string   `json:"expand2"`         // 扩展地段2
	DataVersion     string   `json:"dataVersion"`     // 数据版本
}

// BlockchainData 区块信息模型
type BlockchainData struct {
	TxId           string      `json:"txId"`        // 交易ID
	TxHash         string      `json:"txHash"`      //交易请求hash
	BlockHash      string      `json:"blockHash"`   //当前区块hash
	BlockHeight    uint64      `json:"blockHeight"` //当前区块高度
	Bidbond        string      `json:"bidbond"`     // 投标保函唯一编号
	Bid            string      `json:"bid"`         // 开立投标保函对应招标需求唯一编号
	Progress       string      `json:"progress"`    //进度描述信息
	CreateBy       string      `json:"createBy"`
	CreateTime     uint64      `json:"createTime"`
	Sender         string      `json:"sender"`
	Receiver       []string    `json:"receiver"`
	LastUpdateTime uint64      `json:"lastUpdateTime"`
	LastUpdateBy   string      `json:"lastUpdateBy"`
	BlockData      string      `json:"blockData"`
	Remark         string      `json:"remark"`
	Status         StateEntity `json:"status"` // 状态
}

// StateEntity  状态存储结构(json)
type StateEntity struct {
	ChangeEvent string `json:"changeEvent"` // 变更事件
	PreState    string `json:"preState"`    // 上一状态如"cancel"
	CurrState   string `json:"currState"`   // 当前状态如"applied"
}

type Events struct {
	ChaincodeId  string      `json:"chaincodeId"` //链码ID
	TxId         string      `json:"txId"`        //交易ID
	EventName    string      `json:"eventName"`   //事件名
	Payload      interface{} `json:"payload"`     
}

// BlockchainData 区块信息模型
type BlockDataObj struct {
	BlockHash      string      `json:"blockHash"`   //当前区块hash
	BlockHeight    uint64      `json:"blockHeight"` //当前区块高度
	PreviousHash   string      `json:"previousHash"` //前一个区块哈希
	Events	       interface{} `json:"events"`      //生成的事件结构
}
