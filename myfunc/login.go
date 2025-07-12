package myfunc

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

const jwtSecret = "abcdefg"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	//承载登陆请求中的登陆信息ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}
	//负载部分数据，MapClaims存储Jwt的声明
	//jwt.MapClaims: map[string]interface的一种别名，拥有灵活定义JWT载荷的数据
	if req.Username == "damsu" && req.Password == "123456" {
		claims := jwt.MapClaims{
			"user_id": 1, //id声明
			"exp":     time.Now().Add(time.Hour * 1).Unix(),
		}
		//初始化Jwt对象  指定签名算法，以及jwt的声明。声明编码进入payload

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		//token包含生成完整jwt组件的所有信息
		//构成完整的JWT组件
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "生成命令失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": tokenString})

	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或者密码错误"})
	}
}

// 中间件
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization")
		if tokenString == "" {
			c.JSON(400, gin.H{"error": "未提供令牌"})
			c.Abort()
			return
		}

		//去除bearer前缀
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		//回调函数提供密钥
		//解析jwt格式的字符串对象，转化为jwt对象
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "无效的令牌"})
			c.Abort()
			return
		}
		c.Next()
	}
	//接受一个匿名函数
}

//eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTI5Mzg2MjgsInVzZXJfaWQiOjF9.uZxeUbeoheMQCC9I2SIzVMRiRjn9OlEzZHl0ZHJ_b7Y
