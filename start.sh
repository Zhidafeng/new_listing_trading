#!/bin/bash

# 启动脚本 start.sh
# 用于启动新币交易监控服务

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 配置
APP_NAME="new_listing_trade"
BINARY_NAME="server"
PID_FILE="logs/${APP_NAME}.pid"
LOG_FILE="logs/${APP_NAME}.log"
CONFIG_FILE="config.yaml"
PORT="${PORT:-8080}"

# 检查是否已经运行
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if ps -p "$PID" > /dev/null 2>&1; then
        echo "错误: 服务已在运行中 (PID: $PID)"
        exit 1
    else
        echo "警告: PID文件存在但进程不存在，删除PID文件"
        rm -f "$PID_FILE"
    fi
fi

# 创建日志目录
mkdir -p logs

# 编译程序（如果需要）
if [ ! -f "$BINARY_NAME" ]; then
    echo "正在编译程序..."
    go build -o "$BINARY_NAME" ./cmd/server/main.go
    if [ $? -ne 0 ]; then
        echo "错误: 编译失败"
        exit 1
    fi
fi

# 启动服务
echo "正在启动服务..."
nohup ./"$BINARY_NAME" -config "$CONFIG_FILE" -port "$PORT" >> "$LOG_FILE" 2>&1 &
PID=$!

# 保存PID
echo $PID > "$PID_FILE"

# 等待一下，检查进程是否启动成功
sleep 2
if ps -p "$PID" > /dev/null 2>&1; then
    echo "服务启动成功!"
    echo "  PID: $PID"
    echo "  日志文件: $LOG_FILE"
    echo "  PID文件: $PID_FILE"
    echo "  使用 './stop.sh' 停止服务"
else
    echo "错误: 服务启动失败，请查看日志: $LOG_FILE"
    rm -f "$PID_FILE"
    exit 1
fi

