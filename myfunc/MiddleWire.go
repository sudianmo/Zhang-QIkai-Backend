package myfunc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"time"
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

//func CorsMiddleWire() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		allowOringin := []string{
//			"https://www.Porhub.com",
//			"https://mihayou.com",
//			"http://localhost:8080", // 添加本地开发环境
//			"http://127.0.0.1:8080",
//		}
//		theOrigin := c.Request.Header.Get("Origin")
//		originIsAllowed := false
//		for _, origin := range allowOringin {
//			if origin == theOrigin {
//				originIsAllowed = true
//				break
//			}
//		}
//		if originIsAllowed {
//			c.Header("Access-Control-Allow-Origin", theOrigin)
//		} else {
//			c.AbortWithStatus(http.StatusForbidden) //告诉用户权限不足
//			return
//		}
//		c.Header("Access - Control - Allow - Methods", "GET, POST, PUT, DELETE, OPTIONS")
//		c.Header("Access - Control - Allow - Headers", "Origin, Content - Type, Content - Length, Accept - Encoding, X - CSRF - Token, Authorization")
//		c.Header("Access - Control - Expose - Headers", "Content - Length")
//		c.Header("Access - Control - Allow - Credentials", "true")
//		if c.Request.Method == "OPTIONS" {
//			c.AbortWithStatus(http.StatusNoContent)
//			//只是验证，不需要返回数据
//			return
//		}
//
//	}
//}

func CorsMiddleWire() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 简单修复：允许所有来源（适用于测试环境）
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "false")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func LogMiddleWire() gin.HandlerFunc {

	return func(c *gin.Context) {
		startTime := time.Now()

		clientIP := c.ClientIP()
		method := c.Request.Method
		url := c.Request.URL.String()
		userAgent := c.Request.UserAgent()

		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				log.Printf("读取请求体失败: %v", err)
				requestBody = "[读取失败]"
			} else {
				requestBody = string(bodyBytes)
				// 重建请求体流
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		//响应拦截器
		originalWriter := c.Writer
		responseBuffer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = responseBuffer

		c.Next()

		responseBody := responseBuffer.body.String()
		statusCode := c.Writer.Status()
		duration := time.Since(startTime)

		logEntry := map[string]interface{}{
			"timestamp":     startTime.Format(time.RFC3339),
			"client_ip":     clientIP,
			"method":        method,
			"url":           url,
			"user_agent":    userAgent,
			"request_body":  requestBody,
			"response_body": responseBody,
			"status_code":   statusCode,
			"duration_ms":   duration.Milliseconds(),
		}

		// 输出JSON格式日志（便于日志系统解析）
		if logBytes, err := json.Marshal(logEntry); err == nil {
			fmt.Println(string(logBytes))
		}
		//恢复
		c.Writer = originalWriter
	}

}
