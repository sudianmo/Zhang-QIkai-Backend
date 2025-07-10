package myfunc

import (
	"encoding/json"
	"fmt"

	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetStudents(c *gin.Context) {
	//就是一个缓存建用于在redis中通过建查找
	cacheKey := "students:all"
	ctx := c.Request.Context() // 使用请求上下文

	// 尝试从Redis获取缓存
	//传递请求上下文，例如上下文被取消就可以中断redis
	//get返回的信息包含命令执行的结果，以及可能的错误信息，redis.StringCmd类型
	//result提取实际的字符串数据

	if catched, err := rdb.Get(ctx, cacheKey).Result(); err == nil {

		var students []Student
		//反序列化json为go中的结构体
		//redis取出的是字符串形式的json，只是一堆字符，很难处理
		if err := json.Unmarshal([]byte(catched), &students); err == nil {
			c.JSON(200, students)
			return
		}
	}

	// 数据库查询
	rows, err := db.Query("SELECT id,name,tel,study,created_at,updated_at FROM students")
	if err != nil {
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var s Student

		if err := rows.Scan(&s.ID, &s.Name, &s.Tel, &s.Study, &s.CreatedAt, &s.UpdatedAt); err != nil {
			c.JSON(500, gin.H{"error": "数据解析失败"})
			return
		}
		students = append(students, s)
	}

	if data, err := json.Marshal(students); err == nil {
		rdb.Set(ctx, cacheKey, data, time.Minute*5)
	}

	c.JSON(200, students)
}

func CreateStudent(c *gin.Context) {
	var s Student
	ctx:=c.Request.Context()
	//处理http请求中的数据绑定
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(400, gin.H{"error": "无效的数据格式"})
		return
	}
	//时间处理
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	s.CreatedAt = currentTime
	s.UpdatedAt = currentTime
	//数据库插入操作
	//把?当作数据，不然，传入的东西可梦影响到sql语句，%s传入sql语句是直接拼接
	query := "INSERT INTO students(name,tel,study,created_at,updated_at) VALUES (?,?,?,?,?)"
	result, err := db.Exec(query, s.Name, s.Tel, s.Study, s.CreatedAt, s.UpdatedAt)

	if err != nil {
		c.JSON(400, gin.H{"error": "数据库操作失败"})
		return
	}

	clearStudentsCache()
	id, _ := result.LastInsertId()



	newStudent:=Student{
		ID:        int(id),
		Name: s.Name,
		Tel: s.Tel ,
		Study: s.Study ,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
	newStudentjson,err:=json.Marshal(&newStudent);if err != nil {
		c.JSON(400,gin.H{"error":"序列化学生失败"})
		return
	}
	if err:=rdb.Set(ctx,fmt.Sprintf("student:%d",id),newStudentjson,0).Err(); err != nil {
		c.JSON(400,gin.H{"error":"写入缓存失败"})
		return
	}
	c.JSON(200, gin.H{
		"message": "学生创建成功",
		"id":      id})


	fmt.Println("写入缓存成功")

}

func UpdateStudent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的学生id"})
		return
	}
	var updated Student
	//检查格式
	if err := c.ShouldBindJSON(&updated); err != nil {
		//状态码400,<==>StatusBadRequest 客户端请求语法错误
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}
	//执行更新
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// SQL 修改：更新语句增加 updated_at
	result, err := db.Exec("UPDATE students SET name=?, tel=?, study=?, updated_at=? WHERE id=?",
		updated.Name, updated.Tel, updated.Study, currentTime, id)
	if err != nil {
		c.JSON(500, gin.H{"error": "更新失败"})
		return
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		c.JSON(404, gin.H{"error": "学生不存在"})
		return
	}

	c.JSON(200, gin.H{"message": "更新成功"})
	clearStudentsCache(id) // id 为当前学生id

}

func DeleteStudent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效ID格式"})
		return
	}

	result, err := db.Exec("DELETE FROM students WHERE id=?", id)
	if err != nil {
		c.JSON(500, gin.H{"error": "删除失败"})
		return
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		c.JSON(404, gin.H{"error": "学生不存在"})
		return
	}
	c.JSON(200, gin.H{"message": "删除成功"})
	clearStudentsCache(id) // id 为当前学生id
}
func GetStudentById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的学生ID"})
		return
	}

	ctx := c.Request.Context() // 使用请求上下文
	cacheKey := fmt.Sprintf("student:%d", id)

	// 尝试从Redis获取缓存
	if catched, err := rdb.Get(ctx, cacheKey).Result(); err == nil {
		if catched == "NULL" { // 防止穿透的特殊标记
			c.JSON(404, gin.H{"error": "学生不存在"})
			return
		}

		var student Student
		if err := json.Unmarshal([]byte(catched), &student); err == nil {
			c.JSON(200, student)
			return
		}
	}

	// 数据库查询
	var student Student
	rows, err := db.Query("SELECT id,name,tel,study,created_at,updated_at WHERE id=?", id)
	if err != nil {
		c.JSON(500, "查询失败")
		return
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(&student.ID, &student.Name, &student.Tel, &student.Study, &student.CreatedAt, &student.UpdatedAt)
		if err != nil {
			c.JSON(500, gin.H{"error": "数据解析失败"})
			return
		}
		c.JSON(200, student)
	} else {
		c.JSON(404, gin.H{"error": "学生不存在"})
		return
	}

	if err := rows.Err(); err != nil {
		c.JSON(500, gin.H{"error": "遍历数据出错"})
		return
	}

	data, _ := json.Marshal(student)
	rdb.Set(ctx, cacheKey, data, time.Minute*5)

	c.JSON(200, student)
}

func clearStudentsCache(studentID ...int) {
	//清除缓存

	ctx := context.Background() // 使用全局的context
	rdb.Del(ctx, "students:all")
	if len(studentID) > 0 {
		key := fmt.Sprintf("student:%d", studentID[0])
		rdb.Del(ctx, key)
	}

}
