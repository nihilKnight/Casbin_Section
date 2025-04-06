#!/bin/bash
set -e  # 任何命令失败则退出
set -x  # 打印执行的命令

# 定义颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# 数据库配置
DSN="plc_casbiner:P1c_c45b1N@tcp(localhost:3306)/plc_casbin?charset=utf8mb4"

# 清理函数（测试前重置环境）
cleanup() {
    echo "Cleaning up..."
    go run main.go database reset --dsn "$DSN" || true
}

# 测试初始化数据库
test_init() {
    echo "=== Testing database init ==="
    if go run main.go database init --dsn "$DSN" | grep -q "initialized"; then
        echo -e "${GREEN}PASS: Database init succeeded${NC}"
    else
        echo -e "${RED}FAIL: Database init failed${NC}"
        exit 1
    fi
}

# 测试添加用户角色
test_add_user() {
    echo "=== Testing add user-role ==="
    if go run main.go database add user001 operator --dsn "$DSN" | grep -q "Successfully added"; then
        echo -e "${GREEN}PASS: Add user succeeded${NC}"
    else
        echo -e "${RED}FAIL: Add user failed${NC}"
        exit 1
    fi
}

# 测试权限检查（预期允许）
test_allow_request() {
    echo "=== Testing ALLOW scenario ==="
    output=$(go run main.go backend request user001 device_control w --dsn "$DSN")
    if echo "$output" | grep -q "[ALLOW]"; then
        echo -e "${GREEN}PASS: Allow check succeeded${NC}"
    else
        echo -e "${RED}FAIL: Allow check failed${NC}"
        exit 1
    fi
}

# 测试权限检查（预期拒绝）
test_deny_request() {
    echo "=== Testing DENY scenario ==="
    output=$(go run main.go backend request user001 audit_log ex --dsn "$DSN")
    if echo "$output" | grep -q "[DENY]"; then
        echo -e "${GREEN}PASS: Deny check succeeded${NC}"
    else
        echo -e "${RED}FAIL: Deny check failed${NC}"
        exit 1
    fi
}

# 测试数据库重置
test_reset() {
    echo "=== Testing database reset ==="
    if go run main.go database reset --dsn "$DSN" | grep -q "reset successfully"; then
        echo -e "${GREEN}PASS: Database reset succeeded${NC}"
    else
        echo -e "${RED}FAIL: Database reset failed${NC}"
        exit 1
    fi
}

# 执行测试流程
cleanup
test_init
test_add_user
test_allow_request
test_deny_request
test_reset

echo -e "\n${GREEN}All tests passed!${NC}"
