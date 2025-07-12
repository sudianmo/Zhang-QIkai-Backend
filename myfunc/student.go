package myfunc

import (
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-redis/redis/v8"

	"net/http"

	"github.com/gin-gonic/gin"
)

type Student struct {
	Id    int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name  string `json:"name"`
	Study string `json:"study"`
	Tel   int    `json:"tel"`
}

func CreateStudent(c *gin.Context) {
	var student Student

	ctx := c.Request.Context()

	if err != nil {
		return
	}

	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "输入格式有误"})
		return
	}
	result := GormDB.Create(&student)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器操作有误，新增学生失败"})
		return
	}
	datamessage := fmt.Sprintf("%s %d", student.Name, student.Id)
	c.JSON(http.StatusOK, gin.H{"数据库新增成功": datamessage})
	key := fmt.Sprintf("student:%s", student.Name)
	data, err := json.Marshal(student)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "编码失败"})
	}
	setResult := rdb.Set(ctx, key, string(data), 0)
	if setResult.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "新增redis缓存失败"})
	}
	c.JSON(http.StatusOK, gin.H{"redis缓存新增成功": datamessage})
	if err := UpdateAllStudentsCache(c); err == nil {
		fmt.Println("学生列表缓存已经更新")
	}
	sqlDB, _ := GormDB.DB()
	stats := sqlDB.Stats()
	fmt.Printf("MySQL连接池 - 空闲连接数: %d, 活跃连接数: %d\n", stats.Idle, stats.InUse)
}
func GetStudentByName(c *gin.Context) {
	sqlDB, _ := GormDB.DB()
	stats := sqlDB.Stats()
	fmt.Printf("MySQL连接池 - 空闲连接数: %d, 活跃连接数: %d\n", stats.Idle, stats.InUse)
	name := c.Param("name")
	ctx := c.Request.Context()

	caheResult, err := rdb.Get(ctx, name).Result()
	if err != nil {
		if err != redis.Nil {
			fmt.Println("没有找到这个键")
		} else {
			fmt.Println("该条学生缓存获取失败")
		}
		var myStudents []Student
		result := GormDB.Find(&myStudents)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据获取失败"})
		}
		c.JSON(http.StatusOK, gin.H{"data": myStudents})
		return
	}
	var student Student
	if err := json.Unmarshal([]byte(caheResult), &student); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"data": student})
	key := fmt.Sprintf("student:%s", student.Name)
	data, err := json.Marshal(student)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "编码失败"})
	}
	setResult := rdb.Set(ctx, key, string(data), 0)
	if setResult.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "新增redis缓存失败"})
	}
	fmt.Println("新建缓存")
	if err := UpdateAllStudentsCache(c); err == nil {
		fmt.Println("学生列表缓存已经更新")
	}
}
func GetStudents(c *gin.Context) {
	sqlDB, _ := GormDB.DB()
	stats := sqlDB.Stats()
	fmt.Printf("MySQL连接池 - 空闲连接数: %d, 活跃连接数: %d\n", stats.Idle, stats.InUse)
	var students []Student
	ctx := c.Request.Context()

	cacheKey := "students:all"
	cacheKey, err := rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cacheKey), &students); err == nil {
			c.JSON(http.StatusOK, gin.H{"data": students, "source": "cache"})
		}
		return
	}

	result := GormDB.Find(&students)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据获取失败"})
		return
	}
	UpdateAllStudentsCache(c)
	c.JSON(http.StatusOK, gin.H{"data": students, "source": "database"})

}
func UpdateStudent(c *gin.Context) {
	//url获取的数据是字符串格式
	//如果是需要id这样的int类型，就需要strconv.Atoi将字符串转换为整数
	sqlDB, _ := GormDB.DB()
	stats := sqlDB.Stats()
	fmt.Printf("MySQL连接池 - 空闲连接数: %d, 活跃连接数: %d\n", stats.Idle, stats.InUse)
	name := c.Param("name")
	var student Student

	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据格式有误"})
		//400,代表数据输入的格式有错误
	}

	result := GormDB.Where("name = ?", name).Updates(&student)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取行数失败"})
		return
	}

	//意味着服务器执行过程中没有出错，但是没有找到目标资源，对于客户端来说就好像是资源没有找到
	//StatusbadRequest 常用于客户端的语法有错误
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "数据操作失败，请检查姓名是否输入错误"})
	}
	err = UpdateAllStudentsCache(c)
	if err != nil {
		fmt.Println("更新缓存失败")
	}
	deleteStudentCache(c, name)
	key := fmt.Sprintf("student:%s", student.Name)
	data, err := json.Marshal(student)
	if err != nil {
		fmt.Println("更新单个学生缓存失败")
	}
	rdb.Set(c, key, string(data), 0)

}

func DeleteStudent(c *gin.Context) {
	sqlDB, _ := GormDB.DB()
	stats := sqlDB.Stats()
	fmt.Printf("MySQL连接池 - 空闲连接数: %d, 活跃连接数: %d\n", stats.Idle, stats.InUse)
	name := c.Param("name")
	//sqlStr := "DELETE from students where name=?"
	//result, err := db.Exec(sqlStr, name)
	result := GormDB.Where("name = ?", name).Delete(&Student{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器操作失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "数据操作失败，请检查姓名是否输入错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"操作成功": "删除学生"})
	deleteStudentCache(c, name)
	err = UpdateAllStudentsCache(c)
	if err != nil {
		fmt.Println("更新缓存失败")
	}
}

func UpdateAllStudentsCache(c *gin.Context) error {

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
	err := rdb.Del(ctx, key).Err()
	if err != nil {
		fmt.Println("删除缓存失败：", err)
		return
	}
	fmt.Println("成功删除建：", name)
}
func enqueueCacheUpdate(key string, value interface{}) {
	if MycacheQueue != nil {
		MycacheQueue.Enqueue(cache{key, value})

	} else {
		return
	}
}
