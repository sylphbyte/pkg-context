package context

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	jwtClaimKey = "x:jwt:claim"
)

type IJwtClaim interface {
	TakeId() string
	TakeToken() string
	TakeIssuer() string
	IssuerIs(name string) bool
}

type LogContext interface {
	Info(location, msg string, data any)
	Trace(location, msg string, data any)
	Debug(location, msg string, data any)
	Warn(location, msg string, data any)
	Fatal(location, msg string, data any)
	Panic(location, msg string, data any)
	Error(location, message string, err error, data any)
}

type StorageContext interface {
	ReceiveDB(name string) *gorm.DB
	ReceiveRedis(name string) *redis.Client
}

type DataContext interface {
	Get(key string) (val any, ok bool)
	Set(key string, val any)
}

type Context interface {
	context.Context // 标准 context
	LogContext      // 日志功能
	DataContext     // 数据功能
	TakeHeader() IHeader
	Clone() Context

	TakeLogger() ILogger
	//ReceiveDB(name string) *gorm.DB // 工厂
	//ReceiveRedis(name string) *redis.Client

	//Info(location, msg string, data any)
	//Trace(location, msg string, data any)
	//Debug(location, msg string, data any)
	//Warn(location, msg string, data any)
	//Fatal(location, msg string, data any)
	//Panic(location, msg string, data any)
	//Error(location, message string, err error, data any)

	StoreJwtClaim(claim IJwtClaim)
	JwtClaim() (claim IJwtClaim)

	SendError(title string, err error, fields ...H)
	SendWarning(title string, fields ...H)
	SendSuccess(title string, fields ...H)
	SendInfo(title string, fields ...H)
}

var _ Context = (*DefaultContext)(nil)

func NewDefaultContext(endpoint Endpoint, path string) Context {
	return &DefaultContext{
		Context: context.Background(),
		Header: &Header{
			Endpoint:   endpoint,
			PathVal:    path,
			TraceIdVal: generateTraceId(),
		},
		logger: _loggerManager.Receive(string(endpoint)),
	}
}

type DefaultContext struct {
	context.Context
	mapping sync.Map
	Header  *Header `json:"header"`
	logger  ILogger
	event   *event
}

func (d *DefaultContext) robotHeader() (h H) {
	now := time.Now()
	traceId := d.TakeHeader().TraceId()

	h = H{
		"Mark":    d.Header.MarkVal,
		"TraceId": traceId,
		"Command": fmt.Sprintf("grep %s /wider-logs/%s/%s.%d.*.log", traceId, now.Format("200601/02"), d.Header.Endpoint, now.Hour()),
	}
	return
}

func (d *DefaultContext) recover() {
	if r := recover(); r != nil {
		d.Error("x.DefaultContext.recover", "context error", errors.Errorf("%v", r), H{
			"stack": takeStack(),
		})
	}
}

func (d *DefaultContext) takeEvent() *event {
	if d.event == nil {
		d.event = newEvent()
	}

	return d.event
}

func (d *DefaultContext) On(eventName string, handlers ...EventHandler) {
	d.takeEvent().On(eventName, handlers...)
}

func (d *DefaultContext) OffEvent(eventName string) {
	d.takeEvent().Off(eventName)
}

func (d *DefaultContext) Emit(eventName string, payload interface{}) {
	d.takeEvent().Emit(d, eventName, payload)
}

// AsyncEmit 异步触发事件(不等待)
func (d *DefaultContext) AsyncEmit(eventName string, payload interface{}) {
	d.takeEvent().AsyncEmitNoWait(d, eventName, payload)
}

// AsyncEmitAndWait 异步触发事件并等待完成
func (d *DefaultContext) AsyncEmitAndWait(eventName string, payload interface{}) {
	d.takeEvent().AsyncEmit(d, eventName, payload)
}

// SendError 发送错误消息
func (d *DefaultContext) SendError(title string, err error, fields ...H) {
	SafeGo(d, "x.DefaultContext.SendError", func() {
		if errorRoboter == nil {
			return
		}

		fields = append([]H{d.robotHeader()}, fields...)
		if err1 := errorRoboter.Send(title, err.Error(), fields...); err1 != nil {
			d.Error("x.DefaultContext.SendError", "send failed", err1, H{})
		}
	})
}

// SendWarning 发送警告消息
func (d *DefaultContext) SendWarning(title string, fields ...H) {
	SafeGo(d, "x.DefaultContext.SendWarning", func() {
		if warningRoboter == nil {
			return
		}

		fields = append([]H{d.robotHeader()}, fields...)
		if err := warningRoboter.Send(title, "", fields...); err != nil {
			d.Error("x.DefaultContext.SendWarning", "send failed", err, H{})
		}
	})
}

// SendSuccess 发送成功消息
func (d *DefaultContext) SendSuccess(title string, fields ...H) {
	SafeGo(d, "x.DefaultContext.SendSuccess", func() {
		if successRoboter == nil {
			return
		}

		fields = append([]H{d.robotHeader()}, fields...)
		if err := successRoboter.Send(title, "", fields...); err != nil {
			d.Error("x.DefaultContext.SendSuccess", "send failed", err, H{})
		}
	})
}

// SendInfo 发送信息消息
func (d *DefaultContext) SendInfo(title string, fields ...H) {
	SafeGo(d, "x.DefaultContext.SendInfo", func() {
		if infoRoboter == nil {
			return
		}

		fields = append([]H{d.robotHeader()}, fields...)
		if err := infoRoboter.Send(title, "", fields...); err != nil {
			d.Error("x.DefaultContext.SendInfo", "send failed", err, H{})
		}
	})
}

func (d *DefaultContext) makeLoggerMessage(location string, msg string, data any) (message *LoggerMessage) {
	return &LoggerMessage{
		Header:   d.Header,
		Location: location,
		Message:  msg,
		Data:     data,
	}
}

func (d *DefaultContext) Info(location string, msg string, data any) {
	d.logger.Info(d.makeLoggerMessage(location, msg, data))
}

func (d *DefaultContext) Trace(location string, msg string, data any) {
	d.logger.Trace(d.makeLoggerMessage(location, msg, data))
}

func (d *DefaultContext) Debug(location string, msg string, data any) {
	d.logger.Debug(d.makeLoggerMessage(location, msg, data))
}

func (d *DefaultContext) Warn(location string, msg string, data any) {
	d.logger.Warn(d.makeLoggerMessage(location, msg, data))
}

func (d *DefaultContext) Fatal(location string, msg string, data any) {
	d.logger.Fatal(d.makeLoggerMessage(location, msg, data))
}

func (d *DefaultContext) Panic(location string, msg string, data any) {
	d.logger.Panic(d.makeLoggerMessage(location, msg, data))
}

func (d *DefaultContext) Error(location string, message string, err error, data any) {
	d.logger.Error(d.makeLoggerMessage(location, message, data), err)
}

func (d *DefaultContext) TakeHeader() IHeader {
	return d.Header
}

func (d *DefaultContext) TakeLogger() ILogger {
	return d.logger
}

func (d *DefaultContext) Clone() (ctx Context) {
	ctx = &DefaultContext{
		Context: context.Background(),
		Header:  d.Header.Clone(),
		logger:  d.logger,
	}

	return
}

func (d *DefaultContext) Get(key string) (val any, ok bool) {
	return d.mapping.Load(key)
}

func (d *DefaultContext) Set(key string, val any) {
	// 类型白名单检查
	//switch val.(type) {
	//case string, int, int64, float64, bool, []string, []int, map[string]string:
	//	// 允许的类型
	//default:
	//	// 对于复杂类型，尝试序列化以确保安全
	//	if _, err := _json.Marshal(val); err != nil {
	//		// 记录警告并拒绝不安全的值
	//		d.Warn("x.DefaultContext.Set", "unsafe value type", H{
	//			"key":  key,
	//			"err":  err.Error(),
	//			"type": fmt.Sprintf("%T", val),
	//		})
	//		return
	//	}
	//}

	d.mapping.Store(key, val)
}

func (d *DefaultContext) StoreJwtClaim(claim IJwtClaim) {
	d.mapping.Store(jwtClaimKey, claim)
}

func (d *DefaultContext) JwtClaim() (claim IJwtClaim) {
	val, ok := d.mapping.Load(jwtClaimKey)
	if !ok {
		return
	}

	return val.(IJwtClaim)
}

// 新增统一的错误处理工具函数
func handleError(ctx Context, location string, operation string, err error, data H) {
	if err == nil {
		return
	}

	if data == nil {
		data = H{}
	}

	data["operation"] = operation
	ctx.Error(location, operation+" failed", err, data)
}
