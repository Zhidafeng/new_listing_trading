#!/bin/bash

# 停止脚本 stop.sh
# 用于停止新币交易监控服务

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 配置
APP_NAME="new_listing_trade"
PID_FILE="logs/${APP_NAME}.pid"

# 检查PID文件是否存在
if [ ! -f "$PID_FILE" ]; then
    echo "错误: PID文件不存在，服务可能未运行"
    exit 1
fi

# 读取PID
PID=$(cat "$PID_FILE")

# 检查进程是否存在
if ! ps -p "$PID" > /dev/null 2>&1; then
    echo "警告: 进程不存在 (PID: $PID)，删除PID文件"
    rm -f "$PID_FILE"
    exit 0
fi

# 停止服务
echo "正在停止服务 (PID: $PID)..."
kill "$PID"

# 等待进程结束
for i in {1..30}; do
    if ! ps -p "$PID" > /dev/null 2>&1; then
        echo "服务已停止"
        rm -f "$PID_FILE"
        exit 0
    fi
    sleep 1
done

# 如果还没停止，强制杀死
if ps -p "$PID" > /dev/null 2>&1; then
    echo "警告: 进程未正常退出，强制终止..."
    kill -9 "$PID"
    sleep 1
    if ! ps -p "$PID" > /dev/null 2>&1; then
        echo "服务已强制停止"
        rm -f "$PID_FILE"
        exit 0
    else
        echo "错误: 无法停止服务"
        exit 1
    fi
fi

