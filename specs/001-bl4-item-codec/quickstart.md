# 快速入门指南: BL4 Item Serial Code Codec

**版本**: 1.0.0
**最后更新**: 2025-10-30

## 概述

BL4 Item Serial Code Codec 是一个用于《无主之地4》游戏物品序列码编解码的工具集。它支持将Base85编码的物品序列码转换为人类可读的结构化数据，并提供反向序列化功能。

## 系统要求

- **Go**: 1.21 或更高版本
- **操作系统**: Linux, Windows, macOS
- **内存**: 最小 64MB，推荐 256MB
- **浏览器**: Chrome 90+, Firefox 88+, Safari 14+ (Web界面)

## 安装步骤

### 1. 克隆仓库
```bash
git clone https://github.com/your-org/bl4-item-codec.git
cd bl4-item-codec
```

### 2. 安装依赖
```bash
go mod download
```

### 3. 构建项目
```bash
# 构建Web API服务
go build -o bl4-api ./cmd/api

# 构建桌面GUI应用
go build -o bl4-gui ./cmd/gui

# 构建命令行工具
go build -o bl4-cli ./cmd/cli

# 构建TUI分析工具
go build -o bl4-tui ./cmd/tui
```

## 快速开始

### Web API服务

1. **启动服务**
```bash
./bl4-api
# 服务将在 http://localhost:8080 启动
```

2. **访问Web界面**
打开浏览器访问 `http://localhost:8080`

3. **API使用示例**

**解码序列码**:
```bash
curl -X POST http://localhost:8080/api/v1/decode \
  -H "Content-Type: application/json" \
  -d '{
    "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
    "show_bitstream": false
  }'
```

**编码结构化数据**:
```bash
curl -X POST http://localhost:8080/api/v1/encode \
  -H "Content-Type: application/json" \
  -d '{
    "structured_data": "24, 0, 1, 50| 2, 3379|| {76} {2} {3}"
  }'
```

**批量解码**:
```bash
curl -X POST http://localhost:8080/api/v1/batch/decode \
  -H "Content-Type: application/json" \
  -d '{
    "serial_codes": [
      "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
      "@Ugr$WBm/$!m!X=5&qXxA;nj3OOD#<4R"
    ],
    "show_bitstream": false
  }'
```

### 桌面GUI应用

1. **启动应用**
```bash
./bl4-gui
```

2. **使用方法**
- 在"Base85"输入框中粘贴物品序列码
- 在"Parts"输入框中输入结构化数据
- 系统会实时进行双向转换
- 状态栏显示物品信息和处理结果

### 命令行工具

1. **解码序列码**
```bash
./bl4-cli decode "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"
```

2. **编码结构化数据**
```bash
./bl4-cli encode "24, 0, 1, 50| 2, 3379|| {76} {2} {3}"
```

3. **批量处理**
```bash
# 创建输入文件
echo "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}" > codes.txt
echo "@Ugr$WBm/$!m!X=5&qXxA;nj3OOD#<4R" >> codes.txt

# 批量解码
./bl4-cli batch-decode -i codes.txt -o results.json
```

4. **显示帮助**
```bash
./bl4-cli --help
```

### TUI分析工具

1. **启动TUI**
```bash
./bl4-tui
```

2. **主要功能**
- 主菜单：选择物品、注入存档、比较模式
- 物品选择：从数据库中选择和分析物品
- 数据比较：比较不同物品的属性差异
- 存档注入：将物品注入到游戏存档中

## 数据格式说明

### Base85序列码格式
- **前缀**: 必须以"@U"开头
- **字符集**: 0-9, A-Z, a-z, !#$%&()*+-;<=>?@^_`{/}~
- **示例**: `@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}`

### 结构化数据格式
```
level, type_index, manufacturer_index, rarity| category, value|| {part_index} {part_index:value} {part_index:[value1 value2]}
```

**格式说明**:
- `level`: 物品等级 (1-100)
- `type_index`: 类型索引
- `manufacturer_index`: 制造商索引
- `rarity`: 稀有度值
- `category, value`: 属性类别和值
- `{part_index}`: 空零件
- `{part_index:value}`: 带值的零件
- `{part_index:[value1 value2]}`: 带值列表的零件

**示例**:
```
24, 0, 1, 50| 2, 3379|| {76} {2} {3} {75:4} {54:[3 2 1]}
```

## 性能指标

- **单次解码**: < 1ms (平均 200μs)
- **单次编码**: < 1ms (平均 180μs)
- **批量处理**: 100个项目 < 5秒
- **内存使用**: 原始数据 + 20% 缓冲区
- **并发支持**: 无状态设计，支持高并发

## 错误处理

### 常见错误类型

1. **无效前缀**: 序列码必须以"@U"开头
2. **Base85格式错误**: 包含无效字符
3. **魔术头部不匹配**: 不是有效的物品序列码
4. **数据不足**: 序列码被截断
5. **令牌错误**: 未知的令牌类型
6. **数据溢出**: 数值超出范围

### 错误响应格式
```json
{
  "error": "InvalidFormat",
  "message": "序列码必须以@U开头",
  "timestamp": "2025-10-30T10:30:00Z"
}
```

## 高级功能

### 比特流分析
启用比特流显示可以看到底层的二进制数据：
```bash
curl -X POST http://localhost:8080/api/v1/decode \
  -H "Content-Type: application/json" \
  -d '{
    "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
    "show_bitstream": true
  }'
```

### 随机物品生成
```bash
# 生成随机物品序列码
./bl4-cli generate --type pistol --rarity legendary --count 10
```

### 模式替换
```bash
# 替换物品中的特定模式
./bl4-cli replace --pattern "{76:#}" --range 1-100 --template "pattern_template.txt"
```

## 配置选项

### 环境变量
```bash
# 服务端口
export BL4_PORT=8080

# 日志级别
export BL4_LOG_LEVEL=info

# 批量处理大小限制
export BL4_MAX_BATCH_SIZE=1000

# 启用比特流显示
export BL4_ENABLE_BITSTREAM=true
```

### 配置文件 (config.json)
```json
{
  "server": {
    "port": 8080,
    "read_timeout": "30s",
    "write_timeout": "30s"
  },
  "logging": {
    "level": "info",
    "format": "json"
  },
  "features": {
    "enable_bitstream": true,
    "enable_batch_process": true,
    "max_batch_size": 1000
  }
}
```

## 故障排除

### 常见问题

**Q: 解码失败，显示"无效的Base85格式"**
A: 检查序列码是否完整，确保没有多余的空格或换行符

**Q: 桌面应用无法启动**
A: 确保系统支持GUI应用，Linux用户可能需要安装GTK库

**Q: 批量处理速度慢**
A: 检查系统资源使用情况，考虑减少批量大小

**Q: TUI界面显示异常**
A: 确保终端支持UTF-8编码和256色

### 日志查看
```bash
# 查看API服务日志
tail -f logs/bl4-api.log

# 查看详细错误信息
export BL4_LOG_LEVEL=debug
./bl4-api
```

## 开发指南

### 项目结构
```
bl4-item-codec/
├── cmd/                 # 命令行入口
│   ├── api/            # Web API服务
│   ├── gui/            # 桌面GUI应用
│   ├── cli/            # 命令行工具
│   └── tui/            # TUI分析工具
├── internal/           # 内部包
│   ├── codec/          # 编解码核心逻辑
│   ├── api/            # API处理
│   ├── gui/            # GUI界面
│   └── config/         # 配置管理
├── pkg/                # 公共包
│   ├── bitstream/      # 比特流处理
│   ├── base85/         # Base85编解码
│   └── token/          # 令牌处理
├── web/                # Web前端资源
├── configs/            # 配置文件
└── docs/               # 文档
```

### 运行测试
```bash
# 运行所有测试
go test ./...

# 运行性能测试
go test -bench=. ./pkg/...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 支持与反馈

- **GitHub**: https://github.com/your-org/bl4-item-codec
- **文档**: https://docs.bl4-codec.com
- **问题反馈**: https://github.com/your-org/bl4-item-codec/issues

## 许可证

MIT License - 详见 LICENSE 文件