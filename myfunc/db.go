package myfunc

import (
	"context"
	_ "database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var GormDB *gorm.DB
var err error
var rdb *redis.Client
var MycacheQueue *cacheQueue

type cache struct {
	key   string
	value interface{}
}
type cacheQueue struct {
	workQueue chan cache
	quit      chan bool
	wg        sync.WaitGroup
}

func NewCacheQueue(size int) *cacheQueue {
	newQueue := &cacheQueue{
		workQueue: make(chan cache, size),
		quit:      make(chan bool),
		wg:        sync.WaitGroup{},
	}
	return newQueue
}
func (cq *cacheQueue) startWorkers(workerCount int) {
	for i := 0; i < workerCount; i++ {
		cq.wg.Add(1)
		go func() {
			defer cq.wg.Done()
			for {
				select {
				case item := <-cq.workQueue:
					cq.processWithRetry(item, 3)
				case <-cq.quit:
					return
				}
			}
		}()
	}
}

func (cq *cacheQueue) processWithRetry(item cache, retryCount int) {
	for i := 0; i < retryCount; i++ {
		if err := cq.cacheProcesser(item); err == nil {
			return
		}
		// 优化重试间隔为毫秒级
		time.Sleep(time.Duration(i*200) * time.Millisecond)
	}
	log.Printf("Failed after %d retry for item: %+v", retryCount, item)
}
func (cq *cacheQueue) cacheProcesser(item cache) error {
	ctx := context.Background()
	err := rdb.Set(ctx, item.key, item.value, time.Minute*5).Err()
	if err != nil {
		log.Printf("缓存写入失败 key:%s ,error:%v", item.key, err)
		return err
	}
	log.Printf("缓存写入成功，key: %s", item.key)
	return nil
}
func (cq *cacheQueue) Enqueue(item cache) {
	select {
	case cq.workQueue <- item:
	default:
		log.Println("队列满了，缓存被丢弃")
	}
}
func (cq *cacheQueue) Close() {
	close(cq.quit)
	cq.wg.Wait()
}

func InitRedis() error {
	rdb = redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     25,
		MinIdleConns: 5,
		MaxRetries:   3,
		PoolTimeout:  time.Second * 4,
		IdleTimeout:  time.Minute * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return err
	}
	MycacheQueue = NewCacheQueue(30)
	MycacheQueue.startWorkers(15)
	fmt.Println("Successfully connected to Redis")
	return nil
}
func InitDB() error {
	dsn := "root:31415926@tcp(127.0.0.1:3306)/Student_sql"
	GormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	sqlDB, _ := GormDB.DB()

	// 优化数据库连接池配置
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Minute * 30) // 修复：原来重复设置了MaxLifetime
	fmt.Println("数据库连接池配置- MICaxdleConns;%d ,MaxOpenConns :%d\n,10,100")

	err = sqlDB.Ping()
	if err != nil {
		return err
	}

	if err := GormDB.AutoMigrate(&Student{}); err != nil {
		return err
	}
	fmt.Println("Successfully connected to DB")
	return nil
}
