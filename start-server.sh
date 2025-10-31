#!/bin/bash

# 简单的BL4 API服务器启动脚本

echo "🚀 Starting BL4 Item Codec API Server..."

# 设置环境变量来修复macOS LC_UUID问题
export BL4_ENVIRONMENT=development
export BL4_SERVER_PORT=8080
export BL4_LOG_LEVEL=info
export BL4_LOG_FORMAT=console
export CGO_ENABLED=0
export GOOS=darwin

# Stay in project root and run API
echo "📊 Starting server on port 8080..."
echo "🌐 Web interface will be available at: http://localhost:8080"
echo "📚 API docs available at: http://localhost:8080/api-docs"
echo "❤️ Health check at: http://localhost:8080/health"
echo ""

# 使用CGO_ENABLED=0来避免LC_UUID问题
exec env CGO_ENABLED=0 go run -ldflags="-w -s" ./cmd/api