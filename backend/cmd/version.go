package main

// VERSION 当前应用版本号（语义化版本）
const VERSION = "1.0.0"

// BUILD_TIME 编译时间，由 ldflags 注入
var BUILD_TIME = "unknown"

// GIT_COMMIT Git 提交哈希，由 ldflags 注入
var GIT_COMMIT = "unknown"
