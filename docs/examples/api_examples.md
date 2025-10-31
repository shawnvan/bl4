# BL4 Item Codec API Examples

This document provides comprehensive examples for using the BL4 Item Codec API. All examples include curl commands, request/response payloads, and explanations.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Health Checks](#health-checks)
3. [Item Decoding](#item-decoding)
4. [Item Encoding](#item-encoding)
5. [Validation](#validation)
6. [Batch Operations](#batch-operations)
7. [Error Handling](#error-handling)
8. [Rate Limiting](#rate-limiting)
9. [Advanced Usage](#advanced-usage)

## Getting Started

### Base URLs

- **Development**: `http://localhost:8080`
- **Production**: `https://api.bl4.example.com`

### Quick Test

Test that the API is running:

```bash
curl -X GET http://localhost:8080/health
```

Expected response:

```json
{
  "status": "healthy",
  "timestamp": "2025-01-30T12:00:00Z",
  "uptime": "5m32s",
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
    "readiness": "/health/readiness",
    "liveness": "/health/liveness",
    "decode": "/api/v1/items/decode",
    "encode": "/api/v1/items/encode",
    "validate": "/api/v1/items/validate",
    "batch_decode": "/api/v1/items/batch/decode",
    "batch_validate": "/api/v1/items/validate/batch",
    "docs": "/api/v1/docs"
  },
  "checks": {
    "database": "ok",
    "memory": "ok",
    "cpu": "ok"
  }
}
```

## Health Checks

### Basic Health Check

```bash
curl -X GET http://localhost:8080/health
```

### Readiness Check

```bash
curl -X GET http://localhost:8080/health/readiness
```

### Liveness Check

```bash
curl -X GET http://localhost:8080/health/liveness
```

## Item Decoding

### Basic Decoding

Decode a single BL4 item serial code:

```bash
curl -X POST http://localhost:8080/api/v1/items/decode \
  -H "Content-Type: application/json" \
  -d '{
    "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
    "options": {
      "include_bitstream": true,
      "include_tokens": true,
      "format": "structured",
      "pretty_print": true
    }
  }'
```

Expected response:

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
      },
      {
        "index": 1,
        "type": "grip",
        "name": "Pistol Grip",
        "value": 5678
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
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2025-01-30T12:00:00Z",
    "processing_time_ms": 45,
    "properties": {
      "bitstream_length": 256,
      "token_count": 12,
      "validation_passed": true
    }
  }
}
```

### Decoding with Options

Decode with specific options:

```bash
curl -X POST http://localhost:8080/api/v1/items/decode \
  -H "Content-Type: application/json" \
  -d '{
    "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
    "options": {
      "format": "json",
      "include_stats": true,
      "max_parts": 5,
      "timeout": 3000
    }
  }'
```

### Validation-Only Decoding

Validate without full decoding:

```bash
curl -X POST http://localhost:8080/api/v1/items/decode \
  -H "Content-Type: application/json" \
  -d '{
    "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
    "options": {
      "validate_only": true
    }
  }'
```

## Item Encoding

### Basic Encoding

Encode structured item data to a serial code:

```bash
curl -X POST http://localhost:8080/api/v1/items/encode \
  -H "Content-Type: application/json" \
  -d '{
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
        },
        {
          "index": 1,
          "type": "grip",
          "value": 5678
        }
      ],
      "name": "Custom Inferno Pistol",
      "rarity": "legendary"
    },
    "options": {
      "validate_parts": true,
      "include_prefix": true,
      "target_version": "1.0"
    }
  }'
```

Expected response:

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
        "value": 1234,
        "attributes": {
          "element": "fire"
        }
      },
      {
        "index": 1,
        "type": "grip",
        "value": 5678
      }
    ],
    "name": "Custom Inferno Pistol",
    "rarity": "legendary"
  },
  "request_id": "550e8400-e29b-41d4-a716-446655440001",
  "processing_time_ms": 38,
  "metadata": {
    "request_id": "550e8400-e29b-41d4-a716-446655440001",
    "timestamp": "2025-01-30T12:00:01Z",
    "processing_time_ms": 38,
    "properties": {
      "encoded_length": 47,
      "compression_ratio": 0.85,
      "validation_passed": true
    }
  }
}
```

### Encoding with Optimization

Enable size optimization:

```bash
curl -X POST http://localhost:8080/api/v1/items/encode \
  -H "Content-Type: application/json" \
  -d '{
    "item_data": {
      "level": 50,
      "type": "shotgun",
      "manufacturer": "jakobs",
      "parts": [
        {
          "index": 0,
          "type": "barrel",
          "value": 9876
        }
      ]
    },
    "options": {
      "optimize_size": true,
      "validate_parts": true
    }
  }'
```

## Validation

### Single Code Validation

Validate a single serial code:

```bash
curl -X POST http://localhost:8080/api/v1/items/validate \
  -H "Content-Type: application/json" \
  -d '{
    "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
    "options": {
      "strict_mode": true,
      "include_details": true,
      "validate_structure": true
    }
  }'
```

Expected response:

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

### Invalid Code Validation

Test with an invalid serial code:

```bash
curl -X POST http://localhost:8080/api/v1/items/validate \
  -H "Content-Type: application/json" \
  -d '{
    "serial_code": "invalid_code_format"
  }'
```

Expected response:

```json
{
  "success": false,
  "valid": false,
  "serial_code": "invalid_code_format",
  "request_id": "550e8400-e29b-41d4-a716-446655440003",
  "error": "BL4 serial codes must start with '@'",
  "error_type": "format_error",
  "length": 19,
  "base85_length": 0,
  "validation_type": "comprehensive",
  "issues": [
    "Missing @ prefix",
    "contains invalid characters: invalid_code_format"
  ],
  "processing_time_ms": 3
}
```

## Batch Operations

### Batch Decoding

Decode multiple serial codes:

```bash
curl -X POST http://localhost:8080/api/v1/items/batch/decode \
  -H "Content-Type: application/json" \
  -d '{
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
      "per_item_timeout": 5000
    }
  }'
```

Expected response:

```json
{
  "success": true,
  "request_id": "550e8400-e29b-41d4-a716-446655440004",
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
      },
      "request_id": "550e8400-e29b-41d4-a716-446655440005",
      "processing_time_ms": 42
    },
    {
      "success": true,
      "serial_code": "@Bv9kK+3}TYh&$zD%j8N3hYmeN*@}dga!m56$b-rh{11}",
      "item_data": {
        "level": 35,
        "type": "rifle",
        "manufacturer": "vladof"
      },
      "request_id": "550e8400-e29b-41d4-a716-446655440006",
      "processing_time_ms": 38
    },
    {
      "success": true,
      "serial_code": "@Lm8nJ-1}TYi#$xA%k6P4iXldP*@}efb!n78$c-si{22}",
      "item_data": {
        "level": 42,
        "type": "shield",
        "manufacturer": "torgue"
      },
      "request_id": "550e8400-e29b-41d4-a716-446655440007",
      "processing_time_ms": 45
    }
  ],
  "errors": [],
  "processing_time_ms": 156,
  "metadata": {
    "request_id": "550e8400-e29b-41d4-a716-446655440004",
    "timestamp": "2025-01-30T12:00:02Z",
    "processing_time_ms": 156,
    "properties": {
      "batch_size": 3,
      "success_rate": 1.0,
      "average_processing_time_ms": 52
    }
  }
}
```

### Enhanced Batch Decoding with Progress Tracking

Use enhanced batch processing with progress tracking:

```bash
curl -X POST http://localhost:8080/api/v1/items/batch/decode/enhanced \
  -H "Content-Type: application/json" \
  -d '{
    "serial_codes": [
      "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
      "@Bv9kK+3}TYh&$zD%j8N3hYmeN*@}dga!m56$b-rh{11}",
      "@Lm8nJ-1}TYi#$xA%k6P4iXldP*@}efb!n78$c-si{22}"
    ],
    "options": {
      "max_concurrency": 2,
      "report_progress": true
    }
  }'
```

Expected response with enhanced tracking:

```json
{
  "success": true,
  "request_id": "550e8400-e29b-41d4-a716-446655440008",
  "batch_id": "batch_550e8400-e29b-41d4-a716-446655440008",
  "total_codes": 3,
  "successful_codes": 3,
  "failed_codes": 0,
  "results": [...],
  "errors": [],
  "processing_time_ms": 165,
  "progress": {
    "current": 3,
    "total": 3,
    "percentage": 100.0,
    "processing_rate": 18.18,
    "estimated_remaining_ms": 0
  },
  "cancellation_token": "cancel_550e8400-e29b-41d4-a716-446655440008",
  "metadata": {
    "request_id": "550e8400-e29b-41d4-a716-446655440008",
    "timestamp": "2025-01-30T12:00:03Z",
    "processing_time_ms": 165,
    "properties": {
      "batch_id": "batch_550e8400-e29b-41d4-a716-446655440008",
      "enhanced_tracking": true
    }
  }
}
```

### Batch Validation

Validate multiple serial codes:

```bash
curl -X POST http://localhost:8080/api/v1/items/validate/batch \
  -H "Content-Type: application/json" \
  -d '{
    "serial_codes": [
      "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
      "invalid_code",
      "@Bv9kK+3}TYh&$zD%j8N3hYmeN*@}dga!m56$b-rh{11}"
    ],
    "options": {
      "strict_mode": true,
      "include_details": true
    }
  }'
```

Expected response:

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
      "request_id": "550e8400-e29b-41d4-a716-446655440010",
      "length": 47,
      "base85_length": 46,
      "validation_type": "comprehensive",
      "issues": [],
      "processing_time_ms": 8
    },
    {
      "success": false,
      "valid": false,
      "serial_code": "invalid_code",
      "request_id": "550e8400-e29b-41d4-a716-446655440011",
      "error": "BL4 serial codes must start with '@'",
      "error_type": "format_error",
      "length": 12,
      "base85_length": 0,
      "validation_type": "comprehensive",
      "issues": ["Missing @ prefix"],
      "processing_time_ms": 2
    },
    {
      "success": true,
      "valid": true,
      "serial_code": "@Bv9kK+3}TYh&$zD%j8N3hYmeN*@}dga!m56$b-rh{11}",
      "request_id": "550e8400-e29b-41d4-a716-446655440012",
      "length": 47,
      "base85_length": 46,
      "validation_type": "comprehensive",
      "issues": [],
      "processing_time_ms": 9
    }
  ],
  "processing_time_ms": 25
}
```

## Error Handling

### Common Error Responses

#### 400 Bad Request - Invalid JSON

```bash
curl -X POST http://localhost:8080/api/v1/items/decode \
  -H "Content-Type: application/json" \
  -d 'invalid json'
```

Response:

```json
{
  "success": false,
  "request_id": "550e8400-e29b-41d4-a716-446655440013",
  "error": "Invalid request format",
  "error_type": "validation_error",
  "status_code": 400
}
```

#### 400 Bad Request - Missing Required Field

```bash
curl -X POST http://localhost:8080/api/v1/items/decode \
  -H "Content-Type: application/json" \
  -d '{}'
```

Response:

```json
{
  "success": false,
  "request_id": "550e8400-e29b-41d4-a716-446655440014",
  "error": "serial_code is required",
  "error_type": "validation_error",
  "status_code": 400
}
```

#### 429 Too Many Requests - Rate Limited

```bash
# Send many requests quickly to trigger rate limiting
for i in {1..100}; do
  curl -X POST http://localhost:8080/api/v1/items/decode \
    -H "Content-Type: application/json" \
    -d '{"serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}"}' \
    -w "Status: %{http_code}\n" -o /dev/null -s
done
```

Response when rate limited:

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

## Rate Limiting

The API implements multi-tier rate limiting:

1. **Global Rate Limit**: 60 requests per minute per IP
2. **Endpoint-specific Limits**: Different limits for different endpoints
3. **Batch Processing Limits**: Separate limits for batch operations

### Rate Limit Headers

When rate limits are active, responses include these headers:

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1640995200
X-RateLimit-Retry-After: 30
```

## Advanced Usage

### Custom Headers

Add custom headers for better tracking:

```bash
curl -X POST http://localhost:8080/api/v1/items/decode \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: custom-request-id" \
  -H "X-Client-Version: 1.0.0" \
  -H "User-Agent: MyBL4Client/1.0" \
  -d '{
    "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
    "metadata": {
      "client_name": "MyBL4Client",
      "request_source": "batch_processing_job"
    }
  }'
```

### Timeout Handling

Set appropriate timeouts:

```bash
curl -X POST http://localhost:8080/api/v1/items/decode \
  -H "Content-Type: application/json" \
  --max-time 10 \
  -d '{
    "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
    "options": {
      "timeout": 5000
    }
  }'
```

### Request Metadata

Include metadata for better debugging:

```bash
curl -X POST http://localhost:8080/api/v1/items/decode \
  -H "Content-Type: application/json" \
  -d '{
    "serial_code": "@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}",
    "metadata": {
      "user_id": "user123",
      "session_id": "session456",
      "request_purpose": "inventory_analysis",
      "client_info": {
        "name": "InventoryManager",
        "version": "2.1.0"
      }
    }
  }'
```

## Programming Language Examples

### Python

```python
import requests
import json

# Base URL
BASE_URL = "http://localhost:8080"

# Decode an item
def decode_item(serial_code):
    response = requests.post(
        f"{BASE_URL}/api/v1/items/decode",
        headers={"Content-Type": "application/json"},
        json={
            "serial_code": serial_code,
            "options": {
                "include_bitstream": True,
                "include_tokens": True
            }
        }
    )
    return response.json()

# Batch decode
def batch_decode_items(serial_codes):
    response = requests.post(
        f"{BASE_URL}/api/v1/items/batch/decode",
        headers={"Content-Type": "application/json"},
        json={
            "serial_codes": serial_codes,
            "options": {
                "max_concurrency": 5,
                "fail_fast": False
            }
        }
    )
    return response.json()

# Usage
item_data = decode_item("@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}")
print(json.dumps(item_data, indent=2))
```

### JavaScript

```javascript
// Using fetch API
const BASE_URL = 'http://localhost:8080';

async function decodeItem(serialCode) {
    const response = await fetch(`${BASE_URL}/api/v1/items/decode`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            serial_code: serialCode,
            options: {
                include_bitstream: true,
                include_tokens: true
            }
        })
    });

    return await response.json();
}

// Batch decode
async function batchDecodeItems(serialCodes) {
    const response = await fetch(`${BASE_URL}/api/v1/items/batch/decode`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            serial_codes: serialCodes,
            options: {
                max_concurrency: 3,
                report_progress: true
            }
        })
    });

    return await response.json();
}

// Usage
(async () => {
    try {
        const itemData = await decodeItem('@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}');
        console.log(JSON.stringify(itemData, null, 2));
    } catch (error) {
        console.error('Error:', error);
    }
})();
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type DecodeRequest struct {
    SerialCode string        `json:"serial_code"`
    Options    interface{}   `json:"options,omitempty"`
}

type DecodeResponse struct {
    Success     bool                   `json:"success"`
    SerialCode  string                 `json:"serial_code"`
    ItemData    map[string]interface{} `json:"item_data,omitempty"`
    RequestID   string                 `json:"request_id"`
    Error       string                 `json:"error,omitempty"`
}

func decodeItem(serialCode string) (*DecodeResponse, error) {
    reqBody := DecodeRequest{
        SerialCode: serialCode,
        Options: map[string]interface{}{
            "include_bitstream": true,
            "include_tokens":    true,
        },
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post("http://localhost:8080/api/v1/items/decode",
                          "application/json",
                          bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var result DecodeResponse
    err = json.Unmarshal(body, &result)
    if err != nil {
        return nil, err
    }

    return &result, nil
}

func main() {
    item, err := decodeItem("@Ugy3L+2}TYg%$yC%i7M2gZldO)@}cgb!l34$a-qf{00}")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Success: %t\n", item.Success)
    if item.ItemData != nil {
        fmt.Printf("Item Data: %+v\n", item.ItemData)
    }
}
```

## Best Practices

1. **Always validate input** before sending to the API
2. **Use appropriate timeouts** for your use case
3. **Handle rate limits** gracefully with exponential backoff
4. **Include request metadata** for better debugging
5. **Use batch operations** for processing multiple items
6. **Monitor the health endpoint** for service status
7. **Implement proper error handling** for all API calls
8. **Use the validation endpoint** before attempting to decode invalid codes

## Troubleshooting

### Common Issues

1. **CORS Errors**: Ensure proper CORS headers are set for web applications
2. **Rate Limiting**: Check response headers for rate limit information
3. **Timeout Issues**: Increase timeout values or reduce batch sizes
4. **Invalid Serial Codes**: Use the validation endpoint first
5. **Memory Issues**: Monitor memory usage for large batch operations

### Debug Mode

Enable debug logging by setting the environment variable:

```bash
export LOG_LEVEL=debug
./bin/bl4-api
```

### Health Monitoring

Regularly check the health endpoints:

```bash
# Basic health
curl http://localhost:8080/health

# Readiness (for container orchestration)
curl http://localhost:8080/health/readiness

# Liveness (for container orchestration)
curl http://localhost:8080/health/liveness
```

For more information, visit the interactive API documentation at `http://localhost:8080/api-docs`.