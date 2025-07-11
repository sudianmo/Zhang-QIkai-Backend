package myfunc

import (
	"encoding/json"
	"fmt"

	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func GetStudents(c *gin.Context) {
	cacheKey := "students:all"

	ctx := c.Request.Context() // 使用请求上下文

	// 尝试从Redis获取缓存
	if catched, err := rdb.Get(ctx, cacheKey).Result(); err == nil {
		var students []Student
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
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(400, gin.H{"error": "无效的数据格式"})
		return
	}
	//时间处理
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	s.CreatedAt = currentTime
	s.UpdatedAt = currentTime
	//数据库插入操作
	query := "INSERT INTO students(name,tel,study,created_at,updated_at) VALUES (?,?,?,?,?)"
	result, err := db.Exec(query, s.Name, s.Tel, s.Study, s.CreatedAt, s.UpdatedAt)

	if err != nil {
		c.JSON(400, gin.H{"error": "数据库操作失败"})
		return
	}
	id, _ := result.LastInsertId()

	//写入缓存，缓存5min
	ctx := c.Request.Context()
	CreaKey := fmt.Sprintf("student:%d", id)
	if data, err := json.Marshal(s); err == nil {
		rdb.Set(ctx, CreaKey, data, time.Minute*5)
		fmt.Println("写入缓存")
	}

	c.JSON(200, gin.H{
		"message": "学生创建成功",
		"id":      id})
	fmt.Println("学生创建成功，name:", s.Name)

}

func UpdateStudent(c *gin.Context) {
	//先更新数据库再更新缓存中的数据
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
	clearStudentCache(id)

	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("student:%d", id)
	if data, err := json.Marshal(updated); err == nil {
		rdb.Set(ctx, cacheKey, data, time.Minute*5)
		fmt.Println("写入缓存")
	}
	// id 为当前学生id

	c.JSON(200, gin.H{"message": "更新成功"})

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
	clearStudentCache(id) // id 为当前学生id
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
	err = db.QueryRow("SELECT id,name,tel,study,created_at,updated_at  FROM students WHERE id=?", id).Scan(
		&student.ID,
		&student.Name,
		&student.Tel,
		&student.Study,
		&student.CreatedAt,
		&student.UpdatedAt)
	if err != nil {
		c.JSON(500, "查询失败")
		return
	}

	if data, err := json.Marshal(student); err == nil {
		rdb.Set(ctx, cacheKey, data, time.Minute*5)
		fmt.Println("写入缓存")
	}

	c.JSON(200, student)
}

// 只清理单个学生缓存
func clearStudentCache(id int) {
	ctx := context.Background()
	key := fmt.Sprintf("student:%d", id)
	rdb.Del(ctx, key)
}

// 清理所有学生列表缓存
func clearAllStudentsCache() {
	ctx := context.Background()
	rdb.Del(ctx, "students:all")
}
