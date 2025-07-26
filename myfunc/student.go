package myfunc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-redis/redis/v8"
	"net/http"
)

type Student struct {
	Id    int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name  string `json:"name"`
	Study string `json:"study"`
	Tel   int    `json:"tel"`
}

func CreateStudent(c *gin.Context) {
	var student Student

	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "输入格式有误"})
		return
	}
	result := GormDB.Create(&student)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器操作有误，新增学生失败"})
		return
	}
	// 立即响应，不等待缓存操作
	c.JSON(http.StatusOK, gin.H{
		"message": "学生创建成功",
		"data":    student,
	})

	// 异步更新缓存
	go func() {
		key := fmt.Sprintf("student:%s", student.Name)
		data, _ := json.Marshal(student)
		enqueueCacheUpdate(key, string(data))
		UpdateAllStudentsCache()
	}()
}

func GetStudentByName(c *gin.Context) {
	name := c.Param("name")
	cacheKey := fmt.Sprintf("student:%s", name)
	// 使用Background context避免请求上下文取消导致的问题
	cacheValue, err := rdb.Get(context.Background(), cacheKey).Result()

	if err == nil {
		var student Student
		if err := json.Unmarshal([]byte(cacheValue), &student); err == nil {
			c.JSON(http.StatusOK, gin.H{"data": student, "source": "cache"})
			return
		}
	}
	var student Student
	if err == redis.Nil {
		result := GormDB.Where("name = ?", name).First(&student)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": student, "source": "database"})

		// 异步更新缓存
		go func() {
			data, _ := json.Marshal(student)
			enqueueCacheUpdate(cacheKey, string(data))
		}()
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "缓存服务异常"})
}

func GetStudents(c *gin.Context) {
	cacheKey := "students:all"
	// 使用Background context避免请求上下文取消导致的问题
	cacheValue, err := rdb.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var students []Student
		if err := json.Unmarshal([]byte(cacheValue), &students); err == nil {
			c.JSON(http.StatusOK, gin.H{"data": students, "source": "cache"})
			return
		}
	}
	var students []Student
	result := GormDB.Find(&students)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据获取失败"})
		return
	}
	// 立即响应，不等待缓存操作
	c.JSON(http.StatusOK, gin.H{"data": students, "source": "database"})

	// 异步更新缓存
	go func() {
		studentsCache, _ := json.Marshal(students)
		enqueueCacheUpdate(cacheKey, string(studentsCache))
	}()
}

func UpdateStudent(c *gin.Context) {
	name := c.Param("name")
	var student Student

	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据格式有误"})
		return
	}

	result := GormDB.Where("name = ?", name).Updates(&student)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取行数失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "数据操作失败，请检查姓名是否输入错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": student})

	// 异步更新缓存
	go func() {
		// 删除旧缓存
		oldKey := fmt.Sprintf("student:%s", name)
		rdb.Del(context.Background(), oldKey)

		// 更新新缓存
		newKey := fmt.Sprintf("student:%s", student.Name)
		data, _ := json.Marshal(student)
		enqueueCacheUpdate(newKey, string(data))
		UpdateAllStudentsCache()
	}()
}

func DeleteStudent(c *gin.Context) {
	name := c.Param("name")
	result := GormDB.Where("name = ?", name).Delete(&Student{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器操作失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "数据操作失败，请检查姓名是否输入错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})

	// 异步清理缓存
	go func() {
		key := fmt.Sprintf("student:%s", name)
		rdb.Del(context.Background(), key)
		UpdateAllStudentsCache()
	}()
}

func UpdateAllStudentsCache() error {
	var students []Student
	result := GormDB.Find(&students)
	if result.Error != nil {
		return result.Error
	}
	studentsCache, err := json.Marshal(students)
	if err != nil {
		return err
	}
	key := "students:all"
	enqueueCacheUpdate(key, string(studentsCache))
	return nil
}

func deleteStudentCache(ctx *gin.Context, name string) {
	key := fmt.Sprintf("student:%s", name)
	// 使用Background context避免HTTP上下文问题
	err := rdb.Del(context.Background(), key).Err()
	if err != nil {
		fmt.Println("删除缓存失败：", err)
		return
	}
	fmt.Println("成功删除键：", name)
}

func enqueueCacheUpdate(key string, value interface{}) {
	if MycacheQueue != nil {
		MycacheQueue.Enqueue(cache{key, value})
	}
}
