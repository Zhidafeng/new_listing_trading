# New Listing Trade

一个基于 Go 语言的新币交易监控系统。

## 项目简介

New Listing Trade 是一个用于监控和交易新上市加密货币的工具。

## 功能特性

- 监控新上市的加密货币
- 自动交易功能
- 实时价格监控
- 风险控制

## 技术栈

- Go 1.18+
- 其他依赖待添加

## 项目结构

```
.
├── cmd/                    # 主程序入口
│   └── server/            # 服务器入口
├── internal/              # 内部包
│   ├── api/              # API 相关代码
│   ├── service/          # 业务逻辑
│   └── models/           # 数据模型
├── pkg/                   # 可导出的包
│   └── utils/            # 工具函数
├── config.yaml           # 配置文件
├── go.mod                # Go 模块定义
└── README.md             # 项目说明
```

## 快速开始

### 环境要求

- Go 1.18 或更高版本

### 安装依赖

```bash
go mod tidy
```

### 运行

```bash
go run cmd/server/main.go
```

## 配置说明

配置文件位于 `config.yaml`（待创建）

## 开发计划

- [ ] 实现新币监控功能
- [ ] 实现自动交易功能
- [ ] 添加风险控制机制
- [ ] 完善日志和监控

## 许可证

MIT

