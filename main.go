package main

import (
	"GIn_Homework/myfunc"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	r := gin.Default()

	//初始化数据库
	dsn := "root:31415926@tcp(127.0.0.1:3306)/Student_sql"
	fmt.Println("等待连接中")
	if err := myfunc.InitDB(dsn); err != nil {
		log.Println("数据库连接失败", err)
		return
	}

	api := r.Group("/api")
	{
		api.POST("/students", myfunc.CreateStudent)
		api.GET("/students", myfunc.GetStudents)
		api.GET("/students/:id", myfunc.GetStudentById)
		api.PUT("/students/:id", myfunc.UpdateStudent)
		api.DELETE("/students/:id", myfunc.DeleteStudent)
	}

	r.Run()
}
