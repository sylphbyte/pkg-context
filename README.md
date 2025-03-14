# pkg/sylph/x 包

这个包提供了一系列工具和基础设施，用于构建可靠、高性能、并发安全的应用程序。

## 主要组件

### 1. Context

Context 是一个核心接口，提供了日志、存储、事件和数据共享等能力。它继承了标准库的 `context.Context`，并添加了更多功能。

```go
// 创建一个默认上下文
ctx := NewDefaultContext(endpoint, path, storage)

// 记录日志
ctx.Info("location", "message", data)

// 获取数据库连接
db := ctx.ReceiveDB("db_name")

// 发送通知
ctx.SendError("出错了", err, x.H{"key": "value"})
```

### 2. 异步日志

支持异步日志处理，减少日志I/O对主业务的影响。

```go
// 启用异步日志
_loggerManager.EnableAsync(1000) // 缓冲区大小为1000

// 禁用异步日志
_loggerManager.DisableAsync()

// 关闭资源
_loggerManager.Close()
```

### 3. 事件系统

提供基于发布/订阅模式的事件系统。

```go
// 订阅事件
ctx.On("user.created", func(ctcontext.Context, payload interface{}) {
    // 处理事件
})

// 触发事件
ctx.Emit("user.created", userData)

// 异步触发事件(不等待)
ctx.AsyncEmit("user.created", userData)

// 异步触发事件(等待完成)
ctx.AsyncEmitAndWait("user.created", userData)
```

### 4. 机器人通知

提供对多种通知模板的支持。

```go
// 发送错误通知
ctx.SendError("出错了", err, x.H{"details": "some details"})

// 发送成功通知
ctx.SendSuccess("操作成功", x.H{"user": "user123"})
```

## 并发安全

此包中的所有组件都经过并发安全优化，可以在高并发环境下使用。主要优化点包括：

1. 使用 `sync.Mutex` 和 `sync.RWMutex` 保护共享资源
2. 安全的 goroutine 启动和恢复机制
3. 异步操作的等待机制，确保资源能够正确释放
4. 对全局状态的并发访问保护

## 资源管理

提供了完整的资源生命周期管理：

```go
// 关闭并释放所有资源
_loggerManager.Close()
```

## 安全性增强

1. 使用 SHA-256 替代 MD5 进行哈希计算
2. 为 Redis 操作添加超时控制
3. 为敏感操作添加重试和失败处理机制

## 使用建议

1. 尽早初始化全局组件，如 `_loggerManager`
2. 在应用关闭前调用 `Close()` 方法释放资源
3. 使用 `SafeGo` 替代直接使用 `go` 关键字启动 goroutine
4. 利用 `WithContextTimeout` 为I/O操作设置超时

## 常见问题

1. 日志丢失：检查是否启用了异步日志，并且在应用关闭前是否调用了 `Close()`
2. 内存使用过高：考虑调整异步日志的缓冲区大小
3. Redis 连接超时：检查网络设置和 Redis 服务器状态

## 贡献指南

添加新功能时，请确保：

1. 编写单元测试
2. 保证并发安全
3. 实现资源正确释放
4. 添加适当的文档注释 