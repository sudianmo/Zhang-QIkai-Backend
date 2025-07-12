package main

import (
	"GIn_Homework/myfunc"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	fmt.Println("等待连接中...")
	if err := myfunc.InitDB(); err != nil {
		log.Fatal(err)
	}
	sqlDB, err := myfunc.GormDB.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	if err := myfunc.InitRedis(); err != nil {
		log.Fatal(err)
	}
	//连接池不会自动关闭，连接池设计的目的是为了合理复用

	r := gin.Default()
	api := r.Group("/api")
	{
		api.Use(myfunc.CorsMiddleWire(), myfunc.LogMiddleWire())

		api.POST("/students", myfunc.CreateStudent)
		api.GET("/students", myfunc.GetStudents)
		api.GET("/students/:name", myfunc.GetStudentByName)
		api.PUT("/students/:name", myfunc.UpdateStudent)
		api.DELETE("/students/:name", myfunc.DeleteStudent)

	}

	r.Run()
}
