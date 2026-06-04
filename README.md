# Go Blog

一个用Go语言构建的简洁博客系统。

## 功能特点

- 支持Markdown语法
- 响应式设计
- 后台管理界面
- SQLite数据库存储
- 简洁美观的界面

## 技术栈

- Go语言
- Gorilla Mux路由
- SQLite数据库
- Markdown解析
- HTML模板

## 安装和运行

### 1. 克隆项目

```bash
git clone https://github.com/looang/go-blog.git
cd go-blog
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 运行项目

```bash
go run cmd/main.go
```

### 4. 访问网站

- 首页: http://localhost:8080
- 后台管理: http://localhost:8080/admin

## 默认账号

- 用户名: admin
- 密码: admin123

## 项目结构

```
go-blog/
├── cmd/
│   └── main.go              # 主程序入口
├── internal/
│   ├── handler/             # HTTP处理器
│   ├── middleware/          # 中间件
│   ├── model/               # 数据模型
│   └── template/            # HTML模板
├── static/
│   └── css/                 # 样式文件
├── data/                    # 数据库文件
├── go.mod                   # Go模块文件
└── README.md                # 项目说明
```

## API接口

- `GET /api/posts` - 获取所有已发布文章
- `POST /api/posts` - 创建新文章
- `PUT /api/posts/:id` - 更新文章
- `DELETE /api/posts/:id` - 删除文章

## 许可证

MIT License
