# BL4 Item Codec API Documentation

## 概览

BL4 Item Codec API 提供了完整的 RESTful API 用于解码、编码、验证和分析《无主之地4》游戏中的物品序列码。

**基础URL**: `http://localhost:8080`
**API版本**: v1
**数据格式**: JSON
**字符编码**: UTF-8

## 认证

当前版本不需要认证。未来版本可能会添加API密钥认证。

## 请求格式

### 通用请求头

```http
Content-Type: application/json
Accept: application/json
User-Agent: YourApp/1.0
X-Request-ID: optional-request-id
```

### 错误响应格式

```json
{
  "success": false,
  "error": "Error message",
  "error_type": "validation_error",
  "status_code": 400,
  "request_id": "uuid",
  "details": {
    "field": "serial_code",
    "issue": "required field missing"
  }
}
```

## 核心端点

### 1. 物品解码

**POST** `/api/v1/items/decode`

将BL4物品序列码解码为结构化数据。

#### 请求体

```json
{
  "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
  "options": {
    "include_bitstream": true,
    "include_tokens": true,
    "format": "structured",
    "pretty_print": true,
    "validate_only": false,
    "include_stats": true,
    "max_parts": 10,
    "timeout": 5000
  },
  "metadata": {
    "client_name": "MyApp",
    "user_id": "12345"
  }
}
```

#### 响应

```json
{
  "success": true,
  "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
  "item_data": {
    "level": 24,
    "type": "pistol",
    "manufacturer": "maliwan",
    "parts": [
      {
        "index": 0,
        "type": "barrel",
        "name": "Maliwan Elemental Barrel",
        "value": 1234,
        "attributes": {
          "element": "fire",
          "damage": 150
        }
      }
    ],
    "name": "Inferno Pistol",
    "rarity": "legendary",
    "properties": {
      "damage": 2345,
      "accuracy": 87.5,
      "fire_rate": 12.3,
      "magazine_size": 15
    }
  },
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "processing_time_ms": 45,
  "metadata": {
    "bitstream_length": 256,
    "token_count": 12,
    "validation_passed": true
  }
}
```

#### 错误响应

| 状态码 | 错误类型 | 描述 |
|--------|----------|------|
| 400 | validation_error | 无效的请求格式或序列码 |
| 429 | rate_limit_error | 超过速率限制 |
| 500 | processing_error | 内部处理错误 |

### 2. 物品编码

**POST** `/api/v1/items/encode`

将结构化物品数据编码为BL4序列码。

#### 请求体

```json
{
  "item_data": {
    "level": 24,
    "type": "pistol",
    "manufacturer": "maliwan",
    "parts": [
      {
        "index": 0,
        "type": "barrel",
        "value": 1234,
        "attributes": {
          "element": "fire"
        }
      }
    ],
    "name": "Custom Inferno Pistol",
    "rarity": "legendary"
  },
  "options": {
    "validate_parts": true,
    "optimize_size": false,
    "include_prefix": true,
    "target_version": "1.0"
  },
  "metadata": {
    "generation_method": "manual"
  }
}
```

#### 响应

```json
{
  "success": true,
  "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
  "item_data": {
    "level": 24,
    "type": "pistol",
    "manufacturer": "maliwan"
  },
  "request_id": "550e8400-e29b-41d4-a716-446655440001",
  "processing_time_ms": 38,
  "metadata": {
    "encoded_length": 47,
    "compression_ratio": 0.85,
    "validation_passed": true
  }
}
```

### 3. 序列码验证

**POST** `/api/v1/items/validate`

验证序列码的格式和结构正确性。

#### 请求体

```json
{
  "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
  "options": {
    "strict_mode": true,
    "include_details": true,
    "validate_structure": true,
    "timeout": 3000
  },
  "metadata": {
    "validation_context": "user_input"
  }
}
```

#### 响应

```json
{
  "success": true,
  "valid": true,
  "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
  "request_id": "550e8400-e29b-41d4-a716-446655440002",
  "length": 47,
  "base85_length": 46,
  "validation_type": "comprehensive",
  "issues": [],
  "details": {
    "has_prefix": true,
    "base85_valid": true,
    "length_valid": true,
    "structure_valid": true,
    "format_compliant": true
  },
  "processing_time_ms": 12
}
```

## 批量操作

### 1. 批量解码

**POST** `/api/v1/items/batch/decode`

同时解码多个序列码。

#### 请求体

```json
{
  "serial_codes": [
    "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
    "@Bv9kK+3}TYh&$zD%j8N3hYmeN*@}dga!m56$b-rh{11}",
    "@Lm8nJ-1}TYi#$xA%k6P4iXldP*@}efb!n78$c-si{22}"
  ],
  "options": {
    "individual_options": {
      "include_bitstream": false,
      "include_tokens": true,
      "format": "structured"
    },
    "max_concurrency": 3,
    "fail_fast": false,
    "report_progress": true,
    "timeout": 30000,
    "per_item_timeout": 5000,
    "compact_output": false
  },
  "metadata": {
    "batch_id": "batch_001",
    "source": "file_import"
  }
}
```

#### 响应

```json
{
  "success": true,
  "request_id": "550e8400-e29b-41d4-a716-446655440004",
  "batch_id": "batch_550e8400-e29b-41d4-a716-446655440008",
  "total_codes": 3,
  "successful_codes": 3,
  "failed_codes": 0,
  "results": [
    {
      "success": true,
      "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
      "item_data": {
        "level": 24,
        "type": "pistol",
        "manufacturer": "maliwan"
      }
    }
  ],
  "progress": {
    "current": 3,
    "total": 3,
    "percentage": 100.0,
    "processing_rate": 18.18,
    "estimated_remaining_ms": 0
  },
  "errors": [],
  "processing_time_ms": 165,
  "cancellation_token": "cancel_550e8400-e29b-41d4-a716-446655440008"
}
```

### 2. 增强批量解码

**POST** `/api/v1/items/batch/decode/enhanced`

提供实时进度跟踪和取消功能的增强批量解码。

#### 进度查询

**GET** `/api/v1/items/batch/{batch_id}/progress`

#### 响应

```json
{
  "success": true,
  "batch_id": "batch_550e8400-e29b-41d4-a716-446655440008",
  "status": "processing",
  "progress": {
    "current": 15,
    "total": 100,
    "percentage": 15.0,
    "processing_rate": 25.5,
    "estimated_remaining_ms": 3345
  },
  "results": [...],
  "errors": []
}
```

#### 取消操作

**DELETE** `/api/v1/items/batch/{batch_id}/cancel`

### 3. 批量验证

**POST** `/api/v1/items/validate/batch`

同时验证多个序列码。

#### 请求体

```json
{
  "serial_codes": [
    "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
    "invalid_code",
    "@Bv9kK+3}TYh&$zD%j8N3hYmeN*@}dga!m56$b-rh{11}"
  ],
  "options": {
    "strict_mode": true,
    "include_details": true
  },
  "metadata": {
    "validation_batch": "batch_002"
  }
}
```

#### 响应

```json
{
  "success": true,
  "request_id": "550e8400-e29b-41d4-a716-446655440009",
  "total_codes": 3,
  "valid_codes": 2,
  "invalid_codes": 1,
  "results": [
    {
      "success": true,
      "valid": true,
      "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
      "length": 47,
      "base85_length": 46,
      "validation_type": "comprehensive",
      "issues": []
    }
  ],
  "processing_time_ms": 25
}
```

## 分析工具

### 1. 模式分析

**POST** `/api/v1/analysis/pattern`

分析序列码中的模式和统计特征。

#### 请求体

```json
{
  "serial_codes": [
    "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
    "@Bv9kK+3}TYh&$zD%j8N3hYmeN*@}dga!m56$b-rh{11}",
    "@Lm8nJ-1}TYi#$xA%k6P4iXldP*@}efb!n78$c-si{22}"
  ],
  "options": {
    "include_patterns": true,
    "include_statistics": true,
    "include_bitstream": true,
    "include_rarity": true,
    "include_manufacturers": true,
    "include_part_frequency": true,
    "deep_analysis": true,
    "max_patterns": 10,
    "confidence_threshold": 0.7
  }
}
```

#### 响应

```json
{
  "success": true,
  "request_id": "550e8400-e29b-41d4-a716-446655440010",
  "sample_size": 3,
  "patterns": [
    {
      "pattern_name": "Maliwan Elemental",
      "pattern_type": "manufacturer",
      "confidence": 0.85,
      "match_count": 2,
      "total_samples": 3,
      "percentage": 66.7,
      "description": "Maliwan weapons with elemental effects",
      "examples": ["Item 1", "Item 3"]
    }
  ],
  "statistics": {
    "total_items": 3,
    "unique_types": 2,
    "unique_manufacturers": 2,
    "average_level": 35.5,
    "min_level": 24,
    "max_level": 47,
    "average_parts": 3.2,
    "data_quality": 0.92,
    "completeness": 0.88
  },
  "bitstream_analysis": {
    "common_prefixes": ["@Ug", "@Bv"],
    "entropy": 4.2,
    "unique_bytes": 45,
    "total_bytes": 128,
    "average_length": 42.7,
    "repeating_patterns": ["ABC", "XYZ"]
  },
  "rarity_distribution": {
    "legendary": 1,
    "epic": 1,
    "rare": 1
  },
  "manufacturer_distribution": {
    "maliwan": 2,
    "jakobs": 1
  },
  "quality_score": 0.88,
  "confidence": 0.75,
  "insights": [
    "Strong pattern detected: Maliwan Elemental (66.7% of items)",
    "High data quality detected",
    "Common bitstream prefixes detected: 2 patterns"
  ],
  "processing_time": "245ms"
}
```

### 2. 物品生成

**POST** `/api/v1/analysis/generate`

基于游戏规则生成随机物品。

#### 请求体

```json
{
  "count": 5,
  "constraints": {
    "min_level": 20,
    "max_level": 30,
    "types": ["pistol", "rifle"],
    "manufacturers": ["maliwan", "jakobs"],
    "rarities": ["rare", "epic"],
    "min_parts": 2,
    "max_parts": 4,
    "elemental": true
  },
  "options": {
    "include_serial_codes": true,
    "include_metadata": true,
    "format": "json",
    "validate_items": true,
    "realistic_generation": true,
    "seed": 12345
  },
  "metadata": {
    "generation_purpose": "testing",
    "user_preferences": "high_damage"
  }
}
```

#### 响应

```json
{
  "success": true,
  "request_id": "550e8400-e29b-41d4-a716-446655440011",
  "count": 5,
  "items": [
    {
      "item_data": {
        "level": 24,
        "type": "pistol",
        "manufacturer": "maliwan",
        "parts": [
          {
            "index": 0,
            "type": "barrel",
            "name": "Maliwan Elemental Barrel",
            "value": 1234,
            "attributes": {
              "element": "fire"
            }
          }
        ],
        "name": "Elegant Inferno Pistol",
        "rarity": "epic",
        "properties": {
          "damage": 2456,
          "accuracy": 85.2,
          "fire_rate": 11.8,
          "magazine_size": 14,
          "element": "fire",
          "elemental_damage": 368
        }
      },
      "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
      "generation_info": {
        "method": "random",
        "generation_time_ms": 12,
        "seed": 12345,
        "rarity": "epic",
        "part_count": 2,
        "special_properties": ["elemental"]
      },
      "validation": {
        "valid": true,
        "checks": {
          "has_level": true,
          "has_type": true,
          "has_manufacturer": true,
          "valid_level": true
        },
        "score": 1.0
      }
    }
  ],
  "statistics": {
    "level_distribution": {
      "L20": 2,
      "L30": 3
    },
    "type_distribution": {
      "pistol": 3,
      "rifle": 2
    },
    "rarity_distribution": {
      "epic": 3,
      "rare": 2
    },
    "average_parts": 2.8,
    "average_level": 25.2,
    "valid_items": 5,
    "invalid_items": 0,
    "generation_rate": 416.7
  },
  "validation_summary": {
    "total_items": 5,
    "valid_items": 5,
    "invalid_items": 0
  },
  "process_time": "142ms",
  "timestamp": "2025-01-30T12:00:00Z"
}
```

### 3. 组合分析和生成

**POST** `/api/v1/analysis/analyze-and-generate`

先分析输入物品，然后基于分析结果生成新物品。

#### 请求体

```json
{
  "serial_codes": [
    "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
    "@Bv9kK+3}TYh&$zD%j8N3hYmeN*@}dga!m56$b-rh{11}"
  ],
  "generation_params": {
    "count": 3,
    "constraints": {
      "min_level": 20,
      "max_level": 40
    }
  },
  "analysis_options": {
    "include_patterns": true,
    "include_statistics": true
  }
}
```

#### 响应

```json
{
  "success": true,
  "request_id": "550e8400-e29b-41d4-a716-446655440012",
  "analysis": {
    "sample_size": 2,
    "patterns": [
      {
        "pattern_name": "High Level Pistols",
        "confidence": 0.9,
        "percentage": 100.0
      }
    ],
    "statistics": {
      "average_level": 28.5,
      "manufacturer_distribution": {
        "maliwan": 1,
        "jakobs": 1
      }
    },
    "quality_score": 0.85,
    "insights": ["High level items detected"]
  },
  "generation": {
    "count": 3,
    "items": [...],
    "statistics": {
      "average_level": 32.0,
      "valid_items": 3
    }
  },
  "process_time": "380ms",
  "timestamp": "2025-01-30T12:00:00Z"
}
```

### 4. 批量分析

**POST** `/api/v1/analysis/batch`

批量分析多组序列码。

#### 请求体

```json
{
  "batches": [
    {
      "name": "legendary_pistols",
      "serial_codes": ["@code1", "@code2"],
      "options": {
        "include_patterns": true,
        "confidence_threshold": 0.8
      }
    },
    {
      "name": "common_rifles",
      "serial_codes": ["@code3", "@code4", "@code5"],
      "options": {
        "include_statistics": true
      }
    }
  ]
}
```

#### 响应

```json
{
  "success": true,
  "request_id": "550e8400-e29b-41d4-a716-446655440013",
  "batch_count": 2,
  "successful_batches": 2,
  "failed_batches": 0,
  "results": [
    {
      "name": "legendary_pistols",
      "sample_size": 2,
      "patterns": [...],
      "quality_score": 0.92,
      "process_time": "85ms"
    },
    {
      "name": "common_rifles",
      "sample_size": 3,
      "statistics": [...],
      "process_time": "120ms"
    }
  ],
  "process_time": "210ms",
  "timestamp": "2025-01-30T12:00:00Z"
}
```

### 5. 服务统计

**GET** `/api/v1/analysis/stats`

获取分析服务的统计信息。

#### 响应

```json
{
  "success": true,
  "request_id": "550e8400-e29b-41d4-a716-446655440014",
  "stats": {
    "analysis_service": {
      "cache_enabled": true,
      "patterns_loaded": 10,
      "service_status": "active"
    },
    "generator_service": {
      "realistic_mode": true,
      "legendary_enabled": true,
      "service_status": "active"
    },
    "api_info": {
      "version": "1.0.0",
      "timestamp": "2025-01-30T12:00:00Z",
      "endpoints": {
        "pattern_analysis": "/api/v1/analysis/pattern",
        "generate_items": "/api/v1/analysis/generate",
        "analyze_and_generate": "/api/v1/analysis/analyze-and-generate",
        "batch_analysis": "/api/v1/analysis/batch",
        "stats": "/api/v1/analysis/stats"
      }
    }
  }
}
```

## 系统监控

### 健康检查

#### 基本健康检查

**GET** `/health`

```json
{
  "status": "healthy",
  "timestamp": "2025-01-30T12:00:00Z",
  "uptime": "2h30m45s",
  "version": "1.0.0",
  "service": "bl4-item-codec",
  "environment": "development",
  "node": {
    "hostname": "localhost",
    "go_version": "go1.21.0",
    "go_os": "darwin",
    "go_arch": "arm64",
    "num_goroutine": 12
  },
  "memory": {
    "alloc_mb": 15,
    "total_alloc_mb": 45,
    "sys_mb": 78,
    "num_gc": 3
  },
  "endpoints": {
    "health": "/health",
    "decode": "/api/v1/items/decode",
    "encode": "/api/v1/items/encode",
    "validate": "/api/v1/items/validate",
    "batch_decode": "/api/v1/items/batch/decode",
    "batch_validate": "/api/v1/items/validate/batch"
  },
  "checks": {
    "database": "ok",
    "memory": "ok",
    "cpu": "ok"
  }
}
```

#### 就绪检查

**GET** `/health/readiness`

```json
{
  "status": "ready",
  "timestamp": "2025-01-30T12:00:00Z",
  "uptime": "2h30m45s",
  "checks": {
    "initialization": {
      "status": "ok",
      "message": "All components initialized successfully"
    },
    "dependencies": {
      "status": "ok",
      "message": "All dependencies are available"
    },
    "resources": {
      "status": "ok",
      "message": "Memory usage is normal",
      "alloc_mb": 15
    }
  }
}
```

#### 存活检查

**GET** `/health/liveness`

```json
{
  "status": "alive",
  "timestamp": "2025-01-30T12:00:00Z",
  "uptime": "2h30m45s"
}
```

## 速率限制

API实现了多级速率限制：

### 全局限制

- **默认**: 60 请求/分钟/IP
- **突发**: 10 个并发请求

### 批量操作限制

- **批量解码**: 10 请求/分钟
- **批量验证**: 15 请求/分钟
- **模式分析**: 5 请求/分钟

### 限制头

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1640995200
X-RateLimit-Retry-After: 30
```

### 超限响应

```json
{
  "success": false,
  "error": "Rate limit exceeded",
  "error_type": "rate_limit_error",
  "status_code": 429,
  "details": {
    "limit": 60,
    "window": "1m",
    "retry_after": 30
  }
}
```

## CORS支持

API支持跨域资源共享（CORS），配置因环境而异：

### 开发环境

```http
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS, PATCH
Access-Control-Allow-Headers: Origin, Content-Type, Accept, Authorization
Access-Control-Max-Age: 86400
```

### 生产环境

```http
Access-Control-Allow-Origin: https://bl4.example.com
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Origin, Content-Type, Accept, Authorization, X-API-Key
Access-Control-Max-Age: 3600
```

## WebSocket支持（计划中）

未来版本将支持WebSocket连接进行实时更新：

- 批量操作进度更新
- 实时分析结果流
- 系统状态监控

## 错误代码参考

| 错误代码 | HTTP状态码 | 描述 |
|----------|------------|------|
| `validation_error` | 400 | 请求验证失败 |
| `invalid_format` | 400 | 数据格式错误 |
| `empty_input` | 400 | 空输入数据 |
| `invalid_length` | 400 | 数据长度无效 |
| `encoding_error` | 400 | 编码错误 |
| `rate_limit_error` | 429 | 超过速率限制 |
| `processing_error` | 500 | 内部处理错误 |
| `service_unavailable` | 503 | 服务不可用 |
| `timeout_error` | 504 | 请求超时 |

## SDK和客户端库

### Go客户端

```go
import "github.com/shawnvan/bl4/client"

client := client.New("http://localhost:8080")
result, err := client.Decode("@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}")
```

### JavaScript客户端

```javascript
import { BL4Client } from '@bl4/client';

const client = new BL4Client('http://localhost:8080');
const result = await client.decode('@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}');
```

### Python客户端

```python
from bl4_client import BL4Client

client = BL4Client('http://localhost:8080')
result = client.decode('@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}')
```

## 最佳实践

### 1. 错误处理

```javascript
try {
  const response = await fetch('/api/v1/items/decode', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ serial_code: code })
  });

  const data = await response.json();

  if (!data.success) {
    throw new Error(data.error);
  }

  return data.item_data;
} catch (error) {
  console.error('Decode failed:', error);
  throw error;
}
```

### 2. 批量处理

```javascript
// 分批处理大量序列码
const batchSize = 100;
const codes = [...]; // 你的序列码数组

for (let i = 0; i < codes.length; i += batchSize) {
  const batch = codes.slice(i, i + batchSize);
  const result = await decodeBatch(batch);

  // 处理结果...

  // 避免速率限制
  await sleep(1000);
}
```

### 3. 缓存策略

```javascript
// 实现客户端缓存
const cache = new Map();

async function decodeWithCache(serialCode) {
  if (cache.has(serialCode)) {
    return cache.get(serialCode);
  }

  const result = await decode(serialCode);
  cache.set(serialCode, result);

  return result;
}
```

### 4. 重试机制

```javascript
async function retryDecode(serialCode, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await decode(serialCode);
    } catch (error) {
      if (error.status === 429) {
        // 速率限制，等待重试
        const retryAfter = error.headers.get('Retry-After') || 5000;
        await sleep(retryAfter);
        continue;
      }

      if (i === maxRetries - 1) throw error;

      // 指数退避
      await sleep(Math.pow(2, i) * 1000);
    }
  }
}
```

## 版本更新

### v1.0.0 (当前版本)

- ✅ 基础解码和编码功能
- ✅ 批量处理支持
- ✅ 验证功能
- ✅ 模式分析
- ✅ 物品生成
- ✅ RESTful API
- ✅ CORS支持
- ✅ 健康检查

### 计划中的功能

- 🔄 WebSocket实时更新
- 🔄 更多分析算法
- 🔄 机器学习模式识别
- 🔄 高级缓存策略
- 🔄 GraphQL支持

## 支持和反馈

- **文档**: https://docs.bl4.example.com
- **GitHub**: https://github.com/shawnvan/bl4
- **问题反馈**: https://github.com/shawnvan/bl4/issues
- **讨论**: https://github.com/shawnvan/bl4/discussions
- **邮件**: api-support@bl4.example.com

---

*最后更新: 2025-01-30*