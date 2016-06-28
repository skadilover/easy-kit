package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/skadilover/easy-kit/json"
	"github.com/skadilover/easy-kit/log"
	"github.com/skadilover/easy-kit/validate"
)

const (
	StatusReject        = 1403
	StatusInternalError = 1500
	StatusOk            = 0
)

type Rest interface {
	Logger() log.Logger //兼容以前的函数
	Info(format string, v ...interface{})
	Error(format string, v ...interface{})
	Say(format string, v ...interface{})
	SayJson(v interface{})
	SayError(code int, msg string)
	SayToastError(code int, msg string)
	JsonInput() *simplejson.Json
	HttpRequest() *http.Request
	HttpResponseWriter() http.ResponseWriter
}

type httpJsonRest struct {
	l    log.Logger
	r    *http.Request
	w    http.ResponseWriter
	data []byte
	j    *simplejson.Json
}

func (r *httpJsonRest) HttpRequest() *http.Request {
	return r.r
}

func (r *httpJsonRest) HttpResponseWriter() http.ResponseWriter {
	return r.w
}

func (r *httpJsonRest) Logger() log.Logger {
	return r.l
}
func (r *httpJsonRest) JsonInput() *simplejson.Json {
	return r.j
}
func (r *httpJsonRest) Info(format string, v ...interface{}) {
	r.l.Info(format, v...)
}

func (r *httpJsonRest) Error(format string, v ...interface{}) {
	r.l.Error(format, v...)
}

func (r *httpJsonRest) Say(format string, v ...interface{}) {
	fmt.Fprintf(r.w, format, v)
	r.Info("response is:"+format, v)
}

func (r *httpJsonRest) SayJson(v interface{}) {
	data, _ := json.JSONMarshal(v, true)
	fmt.Fprintf(r.w, "%s", data)
	r.Info("response is: %s", data)
}

func (r *httpJsonRest) SayError(code int, msg string) {
	v := make(map[string]interface{})
	v["status"] = code
	v["msg"] = msg
	data, _ := json.JSONMarshal(v, true)
	fmt.Fprintf(r.w, "%s", data)
	r.Info("response is: %s", data)
}

func (r *httpJsonRest) SayToastError(code int, msg string) {
	v := make(map[string]interface{})
	v["status"] = code
	v["user_msg"] = msg
	data, _ := json.JSONMarshal(v, true)
	fmt.Fprintf(r.w, "%s", data)
	r.Info("response is: %s", data)
}

func (r *httpJsonRest) decodeJson() error {
	if j, err := simplejson.NewJson(r.data); err != nil {
		r.Error("create json failed:%s", err.Error())
		return err
	} else {
		r.j = j
		return nil
	}
}

func (r *httpJsonRest) loadParams() error {
	if data, err := ioutil.ReadAll(r.r.Body); err != nil {
		r.Error("Read request body failed:%s", err.Error())
		r.SayError(http.StatusNotAcceptable, "Read request body failed.")
		return err
	} else {
		r.data = data
		r.l.Info("%s", data)
		return nil
	}
}

func (r *httpJsonRest) auth([]byte) error {
	if j, err := simplejson.NewJson(r.data); err != nil {
		r.Error("create json failed:%s", err.Error())
		r.SayError(http.StatusNotAcceptable, "Unmarshal request data failed.")
		return err
	} else {
		r.j = j
		return nil
	}
}

func (r *httpJsonRest) authSchemaByte(schema *validate.Property) error {

	if err := schema.ValidateString(string(r.data)); err != nil {
		return fmt.Errorf("validate failed: %s , body is %s", err.Error(), r.data)
	}
	return nil
}
