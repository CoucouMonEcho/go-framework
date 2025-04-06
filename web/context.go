package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type Context struct {
	Req *http.Request
	// use Resp instead of RespData and RespCode
	// may cause some middleware not working
	Resp http.ResponseWriter

	// for middleware
	RespData []byte
	RespCode int

	MatchedRoute string
	pathParams   map[string]string
	queryParams  url.Values

	templateEngine TemplateEngine
}

func (ctx *Context) Render(templateName string, data any) error {
	var err error
	ctx.RespData, err = ctx.templateEngine.Render(ctx.Req.Context(), templateName, data)
	if err != nil {
		ctx.RespCode = http.StatusInternalServerError
		return err
	}
	// no need other status
	ctx.RespCode = http.StatusOK
	return nil
}

func (c *Context) SetCookie(ck *http.Cookie) {
	http.SetCookie(c.Resp, ck)
}

func (c *Context) RespJSONOK(val any) error {
	return c.RespJSON(http.StatusOK, val)
}

func (c *Context) RespJSON(status int, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	// c.Resp.Header().Set("Content-Type", "application/json")
	// c.Resp.Header().Set("Content-Length", strconv.Itoa(len(data)))
	c.RespCode = status
	c.RespData = data
	return err
}

func (c *Context) BindJSON(val any) error {
	decoder := json.NewDecoder(c.Req.Body)
	// decoder.UseNumber()
	// decoder.DisallowUnknownFields()
	return decoder.Decode(val)
}

func (c *Context) FormValue(key string) *StringValue {
	// form has cache
	err := c.Req.ParseForm()
	if err != nil {
		return &StringValue{
			err: err,
		}
	}
	vals, ok := c.Req.Form[key]
	if !ok {
		return &StringValue{
			err: errors.New("web: key not found"),
		}
	}
	return &StringValue{
		val: vals[0],
	}
}

func (c *Context) QueryValue(key string) *StringValue {
	// query has no cache
	if c.queryParams == nil {
		c.queryParams = c.Req.URL.Query()
	}
	vals, ok := c.queryParams[key]
	if !ok || len(vals) == 0 {
		return &StringValue{
			err: errors.New("web: key not found"),
		}
	}
	return &StringValue{
		val: vals[0],
	}
}

func (c *Context) PathValue(key string) *StringValue {
	val, ok := c.pathParams[key]
	if !ok {
		return &StringValue{
			err: errors.New("web: key not found"),
		}
	}
	return &StringValue{
		val: val,
	}
}

type StringValue struct {
	val string
	err error
}

func (sv *StringValue) String() (string, error) {
	if sv.err != nil {
		return "", sv.err
	}
	return sv.val, nil
}

func (sv *StringValue) AsInt64() (int64, error) {
	if sv.err != nil {
		return 0, sv.err
	}
	return strconv.ParseInt(sv.val, 10, 64)
}

func (sv *StringValue) AsInt32() (int, error) {
	if sv.err != nil {
		return 0, sv.err
	}
	int64Val, err := strconv.ParseInt(sv.val, 10, 32)
	return int(int64Val), err
}

func (sv *StringValue) AsFloat64() (float64, error) {
	if sv.err != nil {
		return 0.0, sv.err
	}
	return strconv.ParseFloat(sv.val, 64)
}

func (sv *StringValue) AsFloat32() (float32, error) {
	if sv.err != nil {
		return 0, sv.err
	}
	float64Val, err := strconv.ParseFloat(sv.val, 32)
	return float32(float64Val), err
}

func (sv *StringValue) AsBool() (bool, error) {
	if sv.err != nil {
		return false, sv.err
	}
	return strconv.ParseBool(sv.val)
}

func (sv *StringValue) AsUint64() (uint64, error) {
	if sv.err != nil {
		return 0, sv.err
	}
	return strconv.ParseUint(sv.val, 10, 64)
}

func (sv *StringValue) AsUint32() (uint32, error) {
	if sv.err != nil {
		return 0, sv.err
	}
	uintVal, err := strconv.ParseUint(sv.val, 10, 32)
	return uint32(uintVal), err
}
