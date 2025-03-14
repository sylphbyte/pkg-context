package context

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

type ITaskName interface {
	Name() string
}

var (
	servId string
)

func init() {
	servId = os.Getenv("LOG_SERV_ID")
}

type Roboter interface {
	Send(text string) error
}

var (
	_json = jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
		UseNumber:              true,
	}.Froze()
)

type IStringer interface {
	String() string
}

type KeyMaker interface {
	Make(key string) string
}

type IObjectRegistry interface {
	Register(name KeyMaker, key string, handler func() interface{})
	Receive(name KeyMaker, key string, share bool) (interface{}, bool)
	MustReceive(name KeyMaker, key string, share bool) interface{}
}

func takeStack() string {
	// 获取程序计数器
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:]) // 跳过当前函数和调用者

	// 创建frames
	frames := runtime.CallersFrames(pcs[:n])

	// 用builder来构建字符串，避免多次内存分配
	var sb strings.Builder

	// 遍历调用栈
	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		// 添加函数名和位置信息
		fmt.Fprintf(&sb, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)

		// 可以设置最大深度限制
		if sb.Len() > 8192 {
			sb.WriteString("...(stack trace too long, truncated)")
			break
		}
	}

	return sb.String()
}

// md5String 使用MD5哈希函数(保留兼容性)
func md5String(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// secureHash 使用SHA-256哈希函数，更安全
func secureHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
