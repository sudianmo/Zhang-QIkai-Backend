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

var GormDB *gorm.DB //定义为全局变量，包级别的变量
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

// quit通道负责发送关闭信号
// wg负责确保所有工作协程完成工作，二者得相互配合。
// sync.waitgrouop重要的作用是阻塞作用，等待所有的携程的任务完成时，在阻塞主携程，直到所有任务完成。
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
		time.Sleep(time.Duration(i+1) * time.Second)
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
		log.Println("队列满了，缓存被丢弃？")
		//高并发模式下，如果所有携程都阻塞等待队列有空间，消耗过多资源
		//牺牲部分数据但是系统能够继续允许
	}

}
func (cq *cacheQueue) Close() {
	close(cq.quit)
	cq.wg.Wait()
}

func InitRedis() error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		PoolSize: 10,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return err
	}
	MycacheQueue = NewCacheQueue(20)
	MycacheQueue.startWorkers(5) //启动五个携程
	fmt.Println("Successfully connected to Redis")
	return nil
}
func InitDB() error {
	dsn := "root:314159@tcp(127.0.0.1:3306)/Student_sql"
	GormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	sqlDB, _ := GormDB.DB()
	if err != nil {
		return err
	}

	fmt.Println("数据库连接池配置- MICaxdleConns;%d ,MaxOpenConns :%d\n,10,100")

	//驱动为Mysql，ds
	err := sqlDB.Ping()
	if err != nil {
		return err
	}

	if err := GormDB.AutoMigrate(&Student{}); err != nil {
		return err
	}
	fmt.Println("Successfully connected to DB")
	return nil

}
