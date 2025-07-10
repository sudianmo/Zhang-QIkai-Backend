package myfunc

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

var db *sql.DB
//连接池变量，go与Mysql 和client进行交互的工具
var rdb *redis.Client
var ctx = context.Background()
//传递请求的上下文信息


type Student struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Tel       string `json:"tel"`
	Study     string `json:"study"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func InitDB(dsn string) error {

	//创建数据库连接
	var err error
	db, err = sql.Open("mysql", dsn)
	//open只是创建连接池，初始化相关的配置和参数
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("数据库连接测试失败：%w", err)
	}
	fmt.Println("数据库连接成功")
		//Ping尝试真正连接数据库



	//Exec函数用于执行部返回结果的sql语句，对数据库插入，修改，删除

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS students (
            id INT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(50) NOT NULL,
            tel VARCHAR(20),
            study VARCHAR(50),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4; 
    `)
	//反引号最常用，避免一些潜在的语法错误
	//返回Client操作对象
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})
	//options配置redis客户端
	//newClient返回的是一个实现了redis接口的对象，rdb用于操作redis

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis连接失败: %w", err)
	}
	fmt.Println("Redis连接成功")

	return nil
}


