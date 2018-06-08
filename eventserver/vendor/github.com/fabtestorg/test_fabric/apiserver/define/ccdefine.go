package define

type InvokeRequest struct {
	Key   string `json:"key"`   //存储数据的key
	Value string `json:"value"` //存储数据的value
}

type InvokeResponse struct {
	TrackId   string         `json:"trackId"`
	ResStatus ResponseStatus `json:"responseStatus"`
	Payload   interface{}    `json:"payload"`
}

type QueryRequest struct {
	DslSyntax string `json:"dslSyntax"` //couchDB 查询语法
	SplitPage Page   `json:"page"`      //分页
	BlockFlag bool   `json:"blockFlag"` //是否区块信息查询
}

type QueryResponse struct {
	TrackId        string         `json:"trackId"`
	ResponseStatus ResponseStatus `json:"responseStatus"`
	Page           Page           `json:"page"`
	Payload        interface{}    `json:"payload"`
}

type ResponseStatus struct {
	StatusCode int    `json:"statusCode"` //错误码0:成功1:失败
	StatusMsg  string `json:"statusMsg"`  //错误信息
}

type Page struct {
	CurrentPage  uint `json:"currentPage"`  //当前页码
	PageSize     uint `json:"pageSize"`     //每个页面显示个数
	TotalRecords uint `json:"totalRecords"` //总记录数
}
