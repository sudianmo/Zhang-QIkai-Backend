# 🚀 Student Management System RESTful API

## 项目概述
这是一个基于Go语言Gin框架开发的学生管理系统API，提供完整的学生管理功能，包括：
1.学生信息的** 创建，读取，更新，删除操作 **
2.Redis缓存加速数据读取响应
3,自动管理创建时间和修改时间
4.清晰的错误处理机制
## ⚙️ 技术栈
| **类别**       | **技术/工具**              |
|----------------|---------------------------|
| **后端框架**   | Gin (Go 语言高性能框架)    |
| **ORM**        | GORM                      |
| **数据库**     | MySQL + 连接池优化          |
| **缓存**       | Redis + 自定义异步缓存队列    |
| **认证**       | JWT (JSON Web Token)      |
| **跨域**       | CORS 中间件                |
| **日志**       | 自定义请求/响应日志中间件      |
| **并发**       | Goroutine + Worker Pool   |
| **开发语言**   | Go 1.24                   |

> 📝 **作者注**：本项目为个人学习练习项目，用于理解和实践以下技术概念：
> - **JWT 认证流程**：Token 生成（HS256 签名）、Bearer 鉴权、中间件拦截验证
> - **Redis 缓存策略**：Cache-Aside 模式、异步写回、自定义缓存队列 + Worker Pool 消费者模式
> - **GORM ORM 操作**：AutoMigrate 自动建表、CRUD 操作、连接池配置
> - **Gin 中间件机制**：CORS 跨域处理、请求/响应日志拦截记录
> - **Go 并发编程**：Goroutine 异步缓存更新、sync.WaitGroup 优雅关闭、Channel 队列通信


## 项目结构


**Zhang-Qikai-Backend/
├── myfunc/                 
│   ├── db.go               
│   │   └── 包含函数:
│   │       - InitDB(dsn string) error
│   │
│   └── student.go          
│       └── 包含函数:
│           - GetStudents(c *gin.Context)
│           - CreateStudent(c *gin.Context)
│           - UpdateStudent(c *gin.Context)
│           - DeleteStudent(c *gin.Context)
│           - GetStudentById(c *gin.Context)
│           - clearStudentsCache()
│
├── README.md            
├── go.mod                 
├── go.sum               
└── main.go                
        - main()**


## API使用方法示例

### 1.创建学生

 POST http://localhost:8080/api/students 
Content-Type: application/json

{
  "name": "周伟宏",
  "tel": "118509142222",
  "study": "计算机科学"
}
成功响应：
{
    "message":"学生创建成功"
    "id":"1"
}

### 2.获取单个学生示例
GET http://localhost:8080/api/students/1 
**成功响应**
json{
    "id": 1,
  "name": "周伟宏",
  "tel": "13800138000",
  "study": "计算机科学",
  "created_at": "2025-07-12 10:30:45",
  "updated_at": "2025-07-12 10:30:45"
}

### 3.更新学生信息
PUT http://localhost:8080/api/students/1 
Content-Type: application/json

{
  "name": "周伟宏",
  "tel": "13900139000",
  "study": "人工智能"
}
**成功响应**
json{
    "message": "更新成功"
}

### 4.删除学生
 DELETE http://localhost:8080/api/students/1 
**成功响应**
json{
     "message": "删除成功"
}

## 核心代码功能
main() 初始化路由和服务器
===============================================================
InitDB() 初始化MySQL和Redis连接
===============================================================
GetStudentById() 通过ID获取单个学生信息
===============================================================
CreateStudent()  创建新学生记录
===============================================================
UpdateStudent()  更新学生信息
===============================================================
DeleteStudent()  删除学生信息
===============================================================
# 🚀 Student Management System RESTful API

## 项目概述
这是一个基于Go语言Gin框架开发的学生管理系统API，提供完整的学生管理功能，包括：
1.学生信息的** 创建，读取，更新，删除操作 **
2.Redis缓存加速数据读取响应
3,自动管理创建时间和修改时间
4.清晰的错误处理机制
## ⚙️ 技术栈
| **类别**       | **技术/工具**              |
|----------------|---------------------------|
| **后端框架**   | Gin (Go 语言高性能框架)    |
| **ORM**        | GORM                      |
| **数据库**     | MySQL + 连接池优化          |
| **缓存**       | Redis + 自定义异步缓存队列    |
| **认证**       | JWT (JSON Web Token)      |
| **跨域**       | CORS 中间件                |
| **日志**       | 自定义请求/响应日志中间件      |
| **并发**       | Goroutine + Worker Pool   |
| **开发语言**   | Go 1.24                   |

> 📝 **作者注**：本项目为个人学习练习项目，用于理解和实践以下技术概念：
> - **JWT 认证流程**：Token 生成（HS256 签名）、Bearer 鉴权、中间件拦截验证
> - **Redis 缓存策略**：Cache-Aside 模式、异步写回、自定义缓存队列 + Worker Pool 消费者模式
> - **GORM ORM 操作**：AutoMigrate 自动建表、CRUD 操作、连接池配置
> - **Gin 中间件机制**：CORS 跨域处理、请求/响应日志拦截记录
> - **Go 并发编程**：Goroutine 异步缓存更新、sync.WaitGroup 优雅关闭、Channel 队列通信


## 项目结构


**Zhang-Qikai-Backend/
├── myfunc/                 
│   ├── db.go               
│   │   └── 包含函数:
│   │       - InitDB(dsn string) error
│   │
│   └── student.go          
│       └── 包含函数:
│           - GetStudents(c *gin.Context)
│           - CreateStudent(c *gin.Context)
│           - UpdateStudent(c *gin.Context)
│           - DeleteStudent(c *gin.Context)
│           - GetStudentById(c *gin.Context)
│           - clearStudentsCache()
│
├── README.md            
├── go.mod                 
├── go.sum               
└── main.go                
        - main()**


## API使用方法示例

### 1.创建学生

 POST http://localhost:8080/api/students 
Content-Type: application/json

{
  "name": "周伟宏",
  "tel": "118509142222",
  "study": "计算机科学"
}
成功响应：
{
    "message":"学生创建成功"
    "id":"1"
}

### 2.获取单个学生示例
GET http://localhost:8080/api/students/1 
**成功响应**
json{
    "id": 1,
  "name": "周伟宏",
  "tel": "13800138000",
  "study": "计算机科学",
  "created_at": "2025-07-12 10:30:45",
  "updated_at": "2025-07-12 10:30:45"
}

### 3.更新学生信息
PUT http://localhost:8080/api/students/1 
Content-Type: application/json

{
  "name": "周伟宏",
  "tel": "13900139000",
  "study": "人工智能"
}
**成功响应**
json{
    "message": "更新成功"
}

### 4.删除学生
 DELETE http://localhost:8080/api/students/1 
**成功响应**
json{
     "message": "删除成功"
}

## 核心代码功能
main() 初始化路由和服务器
===============================================================
InitDB() 初始化MySQL和Redis连接
===============================================================
GetStudentById() 通过ID获取单个学生信息
===============================================================
CreateStudent()  创建新学生记录
===============================================================
UpdateStudent()  更新学生信息
===============================================================
DeleteStudent()  删除学生信息
===============================================================
