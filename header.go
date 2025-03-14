package context

import (
	"github.com/google/uuid"
)

type Endpoint string

type IHeader interface {
	Ref() string // 来源
	StoreRef(ref string)
	Path() string
	StorePath(path string)
	TraceId() string
	WithTraceId(traceId string)
	GenerateTraceId()
	ResetTraceId()
	Mark() string
	WithMark(mark string)
	StoreIP(ip string)
	IP() string
}

type Header struct {
	Endpoint   Endpoint `json:"endpoint"`
	MarkVal    string   `json:"x_mark,omitempty"`
	RefVal     string   `json:"x_ref,omitempty"`  // 来源
	PathVal    string   `json:"x_path,omitempty"` // 当前路径
	TraceIdVal string   `json:"trace_id,omitempty"`

	IPVal string `json:"ip,omitempty"`
}

func NewHeader(endpoint Endpoint) *Header {
	return &Header{
		Endpoint:   endpoint,
		TraceIdVal: generateTraceId(),
	}
}

func (h *Header) Ref() string {
	return h.RefVal
}

func (h *Header) StoreRef(ref string) {
	h.RefVal = ref
}

func (h *Header) Path() string {
	return h.PathVal
}

func (h *Header) StorePath(path string) {
	h.PathVal = path
}

func (h *Header) TraceId() string {
	return h.TraceIdVal
}

func (h *Header) WithTraceId(traceId string) {
	h.TraceIdVal = traceId
}

func (h *Header) GenerateTraceId() {
	if h.TraceIdVal != "" {
		return
	}

	h.TraceIdVal = generateTraceId()
}

func (h *Header) ResetTraceId() {
	h.TraceIdVal = generateTraceId()
}

func (h *Header) Mark() string {
	return h.MarkVal
}

func (h *Header) WithMark(mark string) {
	h.MarkVal = mark
}

func (h *Header) StoreIP(ip string) {
	h.IPVal = ip
}

func (h *Header) IP() string {
	return h.IPVal
}

func (h *Header) Clone() (header *Header) {
	header = &Header{
		Endpoint:   h.Endpoint,
		TraceIdVal: generateTraceId(),
	}

	return
}

func generateTraceId() string {
	return uuid.NewString()
}
