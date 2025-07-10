package myfunc

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var rdb *redis.Client
var ctx = context.Background()

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
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("连接测试失败：%w", err)

	}
	fmt.Println("数据库连接成功")
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
	//返回Client操作对象
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis连接失败%w", err)
	}
	fmt.Println("Redis连接成功")
	return nil
}
