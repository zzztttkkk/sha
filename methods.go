package sha

const (
	MethodGet     = "GET"     // RFC 7231, 4.3.1
	MethodHead    = "HEAD"    // RFC 7231, 4.3.2
	MethodPost    = "POST"    // RFC 7231, 4.3.3
	MethodPut     = "PUT"     // RFC 7231, 4.3.4
	MethodPatch   = "PATCH"   // RFC 5789
	MethodDelete  = "DELETE"  // RFC 7231, 4.3.5
	MethodConnect = "CONNECT" // RFC 7231, 4.3.6
	MethodOptions = "OPTIONS" // RFC 7231, 4.3.7
	MethodTrace   = "TRACE"   // RFC 7231, 4.3.8
)

type _Method int

const (
	_MCustom = _Method(iota)
	_MGet
	_MHead
	_MPost
	_MPut
	_MPatch
	_MDelete
	_MConnect
	_MOptions
	_MTrace
)

func (ctx *RequestCtx) IsGET() bool     { return ctx.Request._method == _MGet }
func (ctx *RequestCtx) IsHEAD() bool    { return ctx.Request._method == _MHead }
func (ctx *RequestCtx) IsPOST() bool    { return ctx.Request._method == _MPost }
func (ctx *RequestCtx) IsPUT() bool     { return ctx.Request._method == _MPut }
func (ctx *RequestCtx) IsPATCH() bool   { return ctx.Request._method == _MPatch }
func (ctx *RequestCtx) IsDELETE() bool  { return ctx.Request._method == _MDelete }
func (ctx *RequestCtx) IsCONNECT() bool { return ctx.Request._method == _MConnect }
func (ctx *RequestCtx) IsOPTIONS() bool { return ctx.Request._method == _MOptions }
func (ctx *RequestCtx) IsTRACE() bool   { return ctx.Request._method == _MTrace }
