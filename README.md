# GoDo

![Go Version](https://img.shields.io/badge/Go-1.24.4-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)

GoDo 是一个基于 Go 语言开发的轻量级任务调度系统，支持 Cron 表达式定时执行，提供完整的 RESTful API 和 Web 管理界面。

## ✨ 特性

- 🕐 **灵活的任务调度** - 支持秒级精度的 Cron 表达式
- 🔐 **用户认证** - 基于 JWT 的用户认证和权限管理
- 📝 **任务管理** - 支持创建、删除、查看任务列表
- 📜 **Shell 脚本执行** - 支持上传和执行 Shell 脚本
- 📊 **任务日志** - 详细的任务执行日志记录
- 🔧 **配置灵活** - 支持 YAML 配置文件
- 📖 **API文档** - 集成 Swagger API 文档
- 🏗️ **依赖注入** - 使用 Google Wire 进行依赖注入
- 📈 **结构化日志** - 基于 Zap 的高性能日志系统

## 🏗️ 项目结构

```
GoDo/
├── api/                    # API 服务层
│   ├── api.go             # HTTP 服务器配置
│   └── route.go           # 路由配置
├── auth/                   # 用户认证
│   └── user.go            # 用户认证逻辑
├── cmd/                    # 应用程序入口
│   ├── main.go            # 主程序
│   └── wire.go            # 依赖注入配置
├── config/                 # 配置管理
│   ├── config.go          # 配置加载
│   ├── config.yaml        # 配置文件
│   └── model.go           # 配置结构体
├── controller/             # 控制器层
│   ├── controller.go      # 基础控制器
│   ├── task.go           # 任务管理控制器
│   └── user.go           # 用户管理控制器
├── dao/                    # 数据访问层
│   ├── dao.go            # 数据库连接
│   ├── taskInfo.go       # 任务信息数据访问
│   ├── taskLog.go        # 任务日志数据访问
│   └── user.go           # 用户数据访问
├── docs/                   # API 文档
├── model/                  # 数据模型
│   ├── taskInfo.go       # 任务信息模型
│   ├── taskLog.go        # 任务日志模型
│   └── user.go           # 用户模型
├── scheduler/              # 任务调度器
│   ├── executor.go       # 任务执行器
│   ├── middleware.go     # 中间件
│   ├── scheduler.go      # 调度器核心
│   ├── task.go          # 任务定义
│   └── job/             # 任务类型
│       ├── job.go       # 任务接口
│       ├── shellJob.go  # Shell 任务实现
│       └── shellJob_test.go
└── test/                   # 测试文件
    └── test.go
```

## 🚀 快速开始

### 环境要求

- Go 1.24.4+
- MySQL 5.7+
- Git

### 安装步骤

1. **克隆项目**
```bash
git clone https://github.com/chencheng8888/GoDo.git
cd GoDo
```

2. **安装依赖**
```bash
go mod download
```


3. **运行程序**
```bash
# 使用默认配置运行
go run cmd/main.go

# 或指定配置文件
go run cmd/main.go -conf config/config.yaml.local
```


## 📖 文档

- API 文档: 启动服务后访问 `/swagger/index.html`
- 配置说明: 参见 [`config/config.yaml`](config/config.yaml)

## 🔧 开发指南

### 构建项目

```bash
# 生成依赖注入代码
go generate ./...

# 构建二进制文件
go build -o bin/godo cmd/main.go

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o bin/godo-linux cmd/main.go
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行指定包的测试
go test ./scheduler/job/

# 运行测试并显示覆盖率
go test -cover ./...
```

### 生成 API 文档

```bash
# 安装 swag
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g cmd/main.go
```

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目基于 MIT 许可证开源 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🐛 问题反馈

如果您发现任何问题或有功能建议，请在 [Issues](https://github.com/chencheng8888/GoDo/issues) 页面提交。