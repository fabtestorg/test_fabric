package utils

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/peersafe/factoring/apiserver/define"
)

// Response http response
func Response(response interface{}, c *gin.Context, status int, responseStatus *define.ResponseStatus, page *define.Page) {
	b, _ := json.Marshal(response)
	c.Writer.Header().Set("version", c.Request.Header.Get("version"))
	c.Writer.Header().Set("content-Type", c.Request.Header.Get("content-Type"))
	c.Writer.Header().Set("trackId", c.Request.Header.Get("trackId"))
	c.Writer.Header().Set("language", c.Request.Header.Get("language"))
	jsonStatus, _ := json.Marshal(responseStatus)
	c.Writer.Header().Set("responseStatus", string(jsonStatus))
	if page != nil {
		jsonPage, _ := json.Marshal(page)
		c.Writer.Header().Set("page", string(jsonPage))
	}

	c.Writer.WriteHeader(status)

	c.Writer.Write(b)
}
