# BL4 Item Serial Code Codec

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/shawnvan/bl4)

**Borderlands 4 Item Serial Code Codec** - 一个高性能的Go语言库，用于解码和编码《无主之地4》游戏中的物品序列码。

## 🚀 功能特性

### 核心功能
- **序列码解码**: 将Base85编码的物品序列码转换为结构化数据
- **数据编码**: 将结构化物品数据编码为有效的BL4序列码
- **格式验证**: 验证序列码格式和结构的正确性
- **批量处理**: 支持大规模批量解码和编码操作
- **实时追踪**: 增强的批量处理支持进度跟踪和取消操作

### 高级分析
- **模式分析**: 识别物品中的常见模式和特征
- **随机生成**: 基于游戏规则生成真实的随机物品
- **统计分析**: 提供详细的统计洞察和数据分布
- **质量评估**: 评估生成物品的质量和真实性

### 多接口支持
- **RESTful API**: 完整的HTTP API服务 (端口8080)
- **桌面GUI**: 跨平台桌面应用程序
- **命令行工具**: 强大的CLI工具用于自动化
- **终端TUI**: 交互式终端界面

## 📋 目录

- [快速开始](#-快速开始)
- [安装](#-安装)
- [使用示例](#-使用示例)
- [API文档](#-api文档)
- [项目结构](#-项目结构)
- [开发指南](#-开发指南)
- [性能基准](#-性能基准)
- [贡献指南](#-贡献指南)
- [许可证](#-许可证)

## 🏃‍♂️ 快速开始

### 前置要求

- Go 1.21 或更高版本
- Git

### 安装

```bash
# 克隆仓库
git clone https://github.com/shawnvan/bl4.git
cd bl4

# 安装依赖
go mod download

# 构建所有组件
make build

# 运行API服务器
./bin/bl4-api

# 或使用Docker
docker-compose up
```

### 基本使用

#### API使用

```bash
# 解码物品序列码
curl -X POST http://localhost:8080/api/v1/items/decode \
  -H "Content-Type: application/json" \
  -d '{"serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"}'

# 编码物品数据
curl -X POST http://localhost:8080/api/v1/items/encode \
  -H "Content-Type: application/json" \
  -d '{"item_data": {"level": 24, "type": "pistol", "manufacturer": "maliwan"}}'

# 批量解码
curl -X POST http://localhost:8080/api/v1/items/batch/decode \
  -H "Content-Type: application/json" \
  -d '{"serial_codes": ["@code1", "@code2", "@code3"]}'
```

#### CLI使用

```bash
# 解码单个序列码
./bin/bl4-cli decode "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"

# 批量处理
./bin/bl4-cli batch --input codes.txt --output results.json

# 生成随机物品
./bin/bl4-cli generate --count 10 --level 50 --type pistol

# 分析模式
./bin/bl4-cli analyze --input codes.txt --patterns
```

## 📦 安装

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/shawnvan/bl4.git
cd bl4

# 安装依赖
go mod tidy

# 构建所有组件
make all

# 或者构建特定组件
make api      # API服务器
make cli      # 命令行工具
make gui      # 桌面应用
make tui      # 终端界面
```

### 二进制下载

从 [Releases](https://github.com/shawnvan/bl4/releases) 页面下载预编译的二进制文件。

### Docker安装

```bash
# 拉取镜像
docker pull ghcr.io/shawnvan/bl4:latest

# 运行API服务器
docker run -p 8080:8080 ghcr.io/shawnvan/bl4:latest

# 使用docker-compose
docker-compose up -d
```

## 💡 使用示例

### 基本解码

```go
package main

import (
    "fmt"
    "log"
    "github.com/shawnvan/bl4/internal/codec/base85"
)

func main() {
    // 创建解码器
    decoder := base85.NewDecoder()

    // 解码序列码
    serialCode := "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"
    b85Part := serialCode[1:]

    data, err := decoder.Decode(b85Part)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Decoded data: %x\n", data)
}
```

### 物品生成

```go
package main

import (
    "fmt"
    "github.com/shawnvan/bl4/internal/api/services"
)

func main() {
    // 创建生成器
    generator := services.NewGeneratorService(nil)

    // 生成请求
    request := &services.GenerationRequest{
        Count: 5,
        Constraints: &services.GenerationConstraints{
            MinLevel:    intPtr(20),
            MaxLevel:    intPtr(30),
            Types:       []string{"pistol", "rifle"},
            Manufacturers: []string{"maliwan", "jakobs"},
        },
        Options: &services.GenerationOptions{
            IncludeSerialCodes: true,
            ValidateItems:      true,
        },
    }

    // 生成物品
    response, err := generator.GenerateItems(request)
    if err != nil {
        panic(err)
    }

    for _, item := range response.Items {
        fmt.Printf("Generated: %s (%s)\n", item.ItemData.Name, item.SerialCode)
    }
}

func intPtr(i int) *int { return &i }
```

### 模式分析

```go
package main

import (
    "fmt"
    "github.com/shawnvan/bl4/internal/api/services"
)

func main() {
    // 创建分析服务
    analysis := services.NewAnalysisService(nil)

    // 分析请求
    request := &services.AnalysisRequest{
        SerialCodes: []string{
            "@code1", "@code2", "@code3", "@code4", "@code5",
        },
        Options: &services.AnalysisOptions{
            IncludePatterns:      true,
            IncludeStatistics:    true,
            IncludeManufacturers: true,
            ConfidenceThreshold:  0.7,
        },
    }

    // 执行分析
    result, err := analysis.AnalyzePatterns(request)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Found %d patterns\n", len(result.Patterns))
    fmt.Printf("Quality score: %.2f\n", result.QualityScore)

    for _, pattern := range result.Patterns {
        fmt.Printf("- %s: %.1f%% confidence\n", pattern.PatternName, pattern.Percentage)
    }
}
```

## 📚 API文档

### REST API端点

#### 核心端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/v1/items/decode` | POST | 解码物品序列码 |
| `/api/v1/items/encode` | POST | 编码物品数据 |
| `/api/v1/items/validate` | POST | 验证序列码 |

#### 批量操作

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/v1/items/batch/decode` | POST | 批量解码 |
| `/api/v1/items/batch/decode/enhanced` | POST | 增强批量解码 |
| `/api/v1/items/validate/batch` | POST | 批量验证 |

#### 分析工具

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/v1/analysis/pattern` | POST | 模式分析 |
| `/api/v1/analysis/generate` | POST | 物品生成 |
| `/api/v1/analysis/analyze-and-generate` | POST | 分析并生成 |
| `/api/v1/analysis/batch` | POST | 批量分析 |
| `/api/v1/analysis/stats` | GET | 服务统计 |

#### 系统监控

| 端点 | 方法 | 描述 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/health/readiness` | GET | 就绪检查 |
| `/health/liveness` | GET | 存活检查 |
| `/api/v1/docs` | GET | API文档 |

### 在线文档

- **Swagger UI**: http://localhost:8080/api-docs
- **API示例**: [docs/examples/api_examples.md](docs/examples/api_examples.md)

## 🏗️ 项目结构

```
bl4/
├── cmd/                    # 应用程序入口点
│   ├── api/               # API服务器
│   ├── cli/               # 命令行工具
│   ├── gui/               # 桌面应用
│   └── tui/               # 终端界面
├── internal/              # 内部包
│   ├── api/              # API层
│   │   ├── handlers/     # HTTP处理器
│   │   ├── middleware/   # 中间件
│   │   ├── models/       # 数据模型
│   │   ├── routes/       # 路由配置
│   │   └── services/     # 业务逻辑
│   ├── codec/            # 编解码核心
│   │   ├── base85/       # Base85编解码
│   │   └── bitstream/    # 比特流处理
│   └── app/              # 应用配置
├── pkg/                   # 公共包
│   ├── logger/           # 日志工具
│   └── validator/        # 验证工具
├── web/                   # Web资源
│   ├── static/          # 静态文件
│   ├── templates/       # HTML模板
│   └── api-docs.html    # API文档
├── docs/                  # 文档
│   ├── examples/        # 示例代码
│   └── api.md          # API文档
├── specs/                 # 规格文档
│   └── 001-bl4-item-codec/
├── tests/                 # 测试文件
│   ├── unit/            # 单元测试
│   ├── integration/     # 集成测试
│   └── fixtures/        # 测试数据
├── scripts/              # 构建脚本
├── docker-compose.yml    # Docker配置
├── Dockerfile           # Docker镜像
├── Makefile            # 构建工具
└── README.md           # 项目说明
```

## 🛠️ 开发指南

### 开发环境设置

```bash
# 克隆仓库
git clone https://github.com/shawnvan/bl4.git
cd bl4

# 安装开发依赖
go mod download
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行测试
make test

# 代码检查
make lint

# 格式化代码
make fmt

# 构建所有组件
make build
```

### 运行测试

```bash
# 运行所有测试
make test

# 运行单元测试
make test-unit

# 运行集成测试
make test-integration

# 生成测试覆盖率报告
make coverage

# 运行基准测试
make bench
```

### 代码贡献

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 代码规范

- 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 进行代码检查
- 编写单元测试，保持高测试覆盖率
- 添加适当的文档注释

## ⚡ 性能基准

### 解码性能

| 操作 | 平均时间 | 吞吐量 |
|------|----------|--------|
| 单个解码 | ~200μs | 5,000 ops/s |
| 批量解码 (100个) | ~15ms | 6,667 ops/s |
| 验证操作 | ~50μs | 20,000 ops/s |

### 内存使用

| 组件 | 平均内存 | 峰值内存 |
|------|----------|----------|
| API服务器 | ~15MB | ~50MB |
| CLI工具 | ~8MB | ~20MB |
| GUI应用 | ~25MB | ~80MB |

### 并发性能

- **API服务器**: 支持1000+并发请求
- **批量处理**: 支持最大1000个项目批量操作
- **内存池**: 使用对象池减少GC压力

## 🔧 配置

### 环境变量

| 变量名 | 默认值 | 描述 |
|--------|--------|------|
| `PORT` | 8080 | API服务器端口 |
| `ENVIRONMENT` | development | 运行环境 |
| `LOG_LEVEL` | info | 日志级别 |
| `CORS_ORIGINS` | * | CORS允许的源 |

### 配置文件

```json
{
  "server": {
    "port": 8080,
    "host": "0.0.0.0",
    "read_timeout": "30s",
    "write_timeout": "30s"
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "stdout"
  },
  "rate_limiting": {
    "enabled": true,
    "requests_per_minute": 60,
    "burst_size": 10
  },
  "features": {
    "enable_caching": true,
    "enable_metrics": true,
    "enable_analysis": true
  }
}
```

## 🐳 Docker部署

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download && go build -o bin/bl4-api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/bl4-api .
COPY --from=builder /app/web ./web

EXPOSE 8080
CMD ["./bl4-api"]
```

### Docker Compose

```yaml
version: '3.8'
services:
  bl4-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=production
      - LOG_LEVEL=info
    volumes:
      - ./config:/app/config
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - bl4-api
    restart: unless-stopped
```

## 🔍 故障排除

### 常见问题

#### Q: 解码失败，返回"invalid Base85 encoding"
A: 确保序列码格式正确，应以`@`开头，后跟有效的Base85字符。

#### Q: API服务器启动失败
A: 检查端口8080是否被占用，或使用`PORT`环境变量指定其他端口。

#### Q: 批量处理速度慢
A: 考虑使用`enhanced`端点进行批量处理，或调整并发参数。

### 调试模式

```bash
# 启用调试日志
export LOG_LEVEL=debug
./bin/bl4-api

# 启用详细模式
./bin/bl4-cli decode --verbose "@code"

# 性能分析
go tool pprof http://localhost:8080/debug/pprof/profile
```

## 📈 监控和指标

### 健康检查

```bash
# 基本健康检查
curl http://localhost:8080/health

# 详细系统信息
curl http://localhost:8080/api/v1/analysis/stats
```

### Prometheus指标

- `bl4_requests_total`: 请求总数
- `bl4_request_duration_seconds`: 请求延迟
- `bl4_batch_operations_total`: 批量操作总数
- `bl4_cache_hits_total`: 缓存命中数

## 🤝 贡献指南

我们欢迎所有形式的贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详细信息。

### 贡献类型

- 🐛 Bug修复
- ✨ 新功能
- 📚 文档改进
- ⚡ 性能优化
- 🧪 测试覆盖

### 开发流程

1. 查看开放的 [Issues](https://github.com/shawnvan/bl4/issues)
2. 创建新的 Issue 讨论功能想法
3. Fork 并创建功能分支
4. 编写代码和测试
5. 提交 Pull Request
6. 代码审查和合并

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- **Borderlands 系列**: Gearbox Software 出色的游戏系列
- **Go社区**: 提供了优秀的编程语言和生态系统
- ** contributors**: 所有为本项目做出贡献的开发者

## 📞 联系方式

- **项目主页**: https://github.com/shawnvan/bl4
- **问题反馈**: https://github.com/shawnvan/bl4/issues
- **讨论**: https://github.com/shawnvan/bl4/discussions
- **邮箱**: dev@bl4.example.com

---

**注意**: 本项目仅供学习和研究使用。请遵守游戏的服务条款和版权法律。