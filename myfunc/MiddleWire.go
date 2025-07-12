package myfunc

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
func (w *responseWriter) WriteString(b string) (int, error) {
	w.body.WriteString(b)
	return w.ResponseWriter.WriteString(b)
}

func CorsMiddleWire() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowOringin := []string{
			"https:www.Porhub.com",
			"https.mihayou.com",
		}
		theOrigin := c.Request.Header.Get("Origin")
		originIsAllowed := false
		for _, origin := range allowOringin {
			if origin == theOrigin {
				originIsAllowed = true
				break
			}
		}
		if originIsAllowed {
			c.Header("Access-Control-Allow-Origin", theOrigin)
		} else {
			c.AbortWithStatus(http.StatusForbidden) //告诉用户权限不足
			return
		}
		c.Header("Access - Control - Allow - Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access - Control - Allow - Headers", "Origin, Content - Type, Content - Length, Accept - Encoding, X - CSRF - Token, Authorization")
		c.Header("Access - Control - Expose - Headers", "Content - Length")
		c.Header("Access - Control - Allow - Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			//只是验证，不需要返回数据
			return
		}

	}
}

func LogMiddleWire() gin.HandlerFunc {
	return func(c *gin.Context) {
		logString := ""
		client := c.ClientIP()

		method := c.Request.Method
		url := c.Request.URL.String()
		logString = fmt.Sprintf("%s %s %s",
			client, method, url)

		//记录请求体
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				//io读完之后流空需要重新赋值
				//中间件不能影响后续
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
		originWriter := c.Writer
		myWriter := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = myWriter
		c.Next()
		requestBody = myWriter.body.String()
		logString += fmt.Sprintf(" %s ", requestBody)
		fmt.Println(logString)
		c.Writer = originWriter
	}
}
