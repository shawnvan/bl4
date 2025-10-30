# Data Model: BL4 Item Serial Code Codec

**Date**: 2025-10-30
**Purpose**: 定义系统核心数据实体和关系

## 核心数据类型

### 1. ItemSerialCode (物品序列码)
```go
type ItemSerialCode struct {
    Code     string `json:"code" validate:"required"`     // Base85编码的序列码
    Prefix   string `json:"prefix"`                       // 前缀标识符 (@U)
    Data     []byte `json:"-"`                             // 解码后的字节数据
    IsValid  bool   `json:"is_valid"`                     // 验证状态
    ErrorMsg string `json:"error_message,omitempty"`      // 错误信息
}
```

### 2. BitStream (比特流)
```go
type BitStream struct {
    Data   []byte `json:"-"`           // 原始字节数据
    Position int   `json:"-"`          // 当前比特位置
    Length   int   `json:"length"`     // 比特流长度
}
```

### 3. Token (令牌)
```go
type TokenType int

const (
    TOK_SEP1   TokenType = iota // 00 - 硬分隔符
    TOK_SEP2                   // 01 - 软分隔符
    TOK_VARINT                 // 100 - 变长整数
    TOK_PART                   // 101 - 复合零件
    TOK_VARBIT                 // 110 - 变长位字段
    TOK_STRING                 // 111 - 字符串
)

type Token struct {
    Type     TokenType `json:"type"`
    Position int       `json:"position"` // 在比特流中的位置
}
```

### 4. VARINT (变长整数)
```go
type VARINT struct {
    Value    uint32 `json:"value"`              // 解码后的数值
    RawBits  []byte `json:"-"`                  // 原始比特数据
    BlockCount int   `json:"block_count"`       // 使用的块数
}
```

### 5. VARBIT (变长位字段)
```go
type VARBIT struct {
    Value   uint32 `json:"value"`        // 解码后的数值
    Length  int    `json:"length"`       // 位长度
    RawBits []byte `json:"-"`           // 原始比特数据
}
```

### 6. PartSubType (零件子类型)
```go
type PartSubType int

const (
    SUBTYPE_NONE PartSubType = iota // 空零件
    SUBTYPE_INT                    // 整数零件
    SUBTYPE_LIST                   // 列表零件
)
```

### 7. PART (复合零件)
```go
type PART struct {
    Index   uint32       `json:"index"`             // 零件索引
    SubType PartSubType  `json:"subtype"`            // 子类型
    Value   uint32       `json:"value,omitempty"`   // 整数值（SUBTYPE_INT）
    Values  []uint32     `json:"values,omitempty"`  // 值列表（SUBTYPE_LIST）
}
```

### 8. B4String (字符串类型)
```go
type B4String struct {
    Value   string `json:"value"`     // 解码后的字符串
    Length  int    `json:"length"`    // 字符串长度
    RawBits []byte `json:"-"`        // 原始比特数据
}
```

### 9. DataBlock (数据块)
```go
type DataBlock struct {
    Token     TokenType `json:"token"`
    Value     uint32    `json:"value,omitempty"`      // VARINT/VARBIT值
    Part      *PART     `json:"part,omitempty"`       // PART数据
    String    *B4String `json:"string,omitempty"`     // STRING数据
    Position  int       `json:"position"`             // 在序列中的位置
}
```

### 10. ItemData (物品数据)
```go
type ItemData struct {
    Level      int                `json:"level,omitempty"`        // 物品等级
    Type       *ItemType          `json:"type,omitempty"`        // 物品类型
    Manufacturer string           `json:"manufacturer,omitempty"` // 制造商
    Rarity     string             `json:"rarity,omitempty"`       // 稀有度
    Parts      []PART             `json:"parts"`                  // 零件列表
    RawBlocks  []DataBlock        `json:"-"`                      // 原始数据块
    Bitstream  *BitStream         `json:"-"`                      // 比特流数据
}
```

### 11. ItemType (物品类型)
```go
type ItemType struct {
    Type         string `json:"type"`           // 类型名称
    Manufacturer string `json:"manufacturer"`   // 制造商
    BaseBarrel   string `json:"base_barrel,omitempty"` // 基础枪管
}
```

## 请求/响应模型

### 1. DecodeRequest (解码请求)
```go
type DecodeRequest struct {
    SerialCode string `json:"serial_code" validate:"required"` // Base85序列码
    ShowBitstream bool `json:"show_bitstream"`                // 是否显示比特流
}
```

### 2. DecodeResponse (解码响应)
```go
type DecodeResponse struct {
    Success     bool      `json:"success"`
    ItemData    *ItemData `json:"item_data,omitempty"`
    Bitstream   string    `json:"bitstream,omitempty"`     // 格式化比特流数据
    ErrorMsg    string    `json:"error_message,omitempty"`
    ProcessTime string    `json:"process_time"`            // 处理时间
}
```

### 3. EncodeRequest (编码请求)
```go
type EncodeRequest struct {
    StructuredData string `json:"structured_data" validate:"required"` // 结构化数据字符串
}
```

### 4. EncodeResponse (编码响应)
```go
type EncodeResponse struct {
    Success     bool   `json:"success"`
    SerialCode  string `json:"serial_code,omitempty"`  // Base85编码结果
    ErrorMsg    string `json:"error_message,omitempty"`
    ProcessTime string `json:"process_time"`           // 处理时间
}
```

### 5. BatchDecodeRequest (批量解码请求)
```go
type BatchDecodeRequest struct {
    SerialCodes []string `json:"serial_codes" validate:"required,min=1,max=1000"`
    ShowBitstream bool   `json:"show_bitstream"`
}
```

### 6. BatchDecodeResponse (批量解码响应)
```go
type BatchDecodeResponse struct {
    Success      bool              `json:"success"`
    Results      []DecodeResult    `json:"results"`
    Errors       []BatchError      `json:"errors,omitempty"`
    ProcessTime  string            `json:"process_time"`
    ProcessedCount int             `json:"processed_count"`
    ErrorCount   int               `json:"error_count"`
}

type DecodeResult struct {
    SerialCode string    `json:"serial_code"`
    ItemData   *ItemData `json:"item_data,omitempty"`
    Bitstream  string    `json:"bitstream,omitempty"`
    ErrorMsg   string    `json:"error_message,omitempty"`
}

type BatchError struct {
    SerialCode string `json:"serial_code"`
    ErrorMsg   string `json:"error_message"`
    Position   int    `json:"position"`
}
```

### 7. BitstreamDisplay (比特流显示)
```go
type BitstreamDisplay struct {
    HexData     string                 `json:"hex_data"`      // 十六进制显示
    BinaryData  string                 `json:"binary_data"`   // 二进制显示
    TokenBreakdown []TokenBreakdown    `json:"token_breakdown"` // 令牌分解
}

type TokenBreakdown struct {
    TokenType string `json:"token_type"`
    Position  int    `json:"position"`
    Length    int    `json:"length"`
    Value     string `json:"value"`
    Description string `json:"description"`
}
```

## 配置模型

### 1. AppConfig (应用配置)
```go
type AppConfig struct {
    Server   ServerConfig   `json:"server"`
    Logging  LoggingConfig  `json:"logging"`
    Features FeaturesConfig `json:"features"`
}

type ServerConfig struct {
    Port         int           `json:"port"`
    ReadTimeout  time.Duration `json:"read_timeout"`
    WriteTimeout time.Duration `json:"write_timeout"`
}

type LoggingConfig struct {
    Level  string `json:"level"`
    Format string `json:"format"`
}

type FeaturesConfig struct {
    EnableBitstream    bool `json:"enable_bitstream"`
    EnableBatchProcess bool `json:"enable_batch_process"`
    MaxBatchSize       int  `json:"max_batch_size"`
}
```

## 数据验证规则

### ItemSerialCode验证
- 必须以"@U"开头
- 只能包含Base85字符集定义的字符
- 最大长度限制：10KB

### 结构化数据验证
- 必须符合格式规范：`level, type, manufacturer|...|| {part} {part}...`
- 零件索引必须为非负整数
- 数值范围必须合理（0-65535）

### 性能约束
- 单次处理时间：<1ms
- 批量处理：最大1000个项目
- 内存使用：不超过原始数据+20%

## 状态转换图

```
[Base85输入] → [Base85解码] → [比特流解析] → [令牌识别] → [数据类型解码] → [ItemData]
     ↑                                                           ↓
[Base85输出] ← [Base85编码] ← [比特流生成] ← [令牌序列化] ← [数据类型编码]
```

## 错误类型定义

```go
type ErrorCode int

const (
    ErrInvalidPrefix      ErrorCode = 1001
    ErrInvalidBase85      ErrorCode = 1002
    ErrInvalidMagicHeader ErrorCode = 1003
    ErrInvalidToken       ErrorCode = 1004
    ErrDataInsufficient   ErrorCode = 1005
    ErrDataOverflow       ErrorCode = 1006
    ErrUnknownPartType    ErrorCode = 1007
    ErrMalformedData      ErrorCode = 1008
)
```

## 性能优化数据结构

### 预计算查找表
```go
type MirrorTables struct {
    Uint2Mirror  [4]byte   // 2位镜像
    Uint3Mirror  [8]byte   // 3位镜像
    Uint4Mirror  [16]byte  // 4位镜像
    Uint5Mirror  [32]byte  // 5位镜像
    Uint7Mirror  [128]byte // 7位镜像
    Uint8Mirror  [256]byte // 8位镜像
    Uint11Mirror [2048]uint32 // 11位镜像
}

type Base85Tables struct {
    EncodeTable [85]byte
    DecodeTable [256]byte // ASCII到Base85的映射
}
```