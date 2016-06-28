package rest

import (
    "net/http"

    "github.com/skadilover/easykit/log"
    "github.com/skadilover/easykit/validate"
)

//users need to realize RestHandler
type RestHandler interface {
    ServeRest(r Rest) error
}

type Validatable interface {
    GetValidateSchema() *validate.Property
}

type httpHandler struct {
    f func(w http.ResponseWriter, r *http.Request)
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    h.f(w, r)
}

const (
    modeJson = 0
    modeForm = 1
)

//定义http请求的基本流程
func getHttpHandler(mode int, h RestHandler, schema *validate.Property) http.Handler {
    return &httpHandler{
        f: func(w http.ResponseWriter, r *http.Request) {
            l := log.GetHttpLogger(r)
            rest := &httpJsonRest{
                l:  l,
                w:  w,
                r:  r,
            }
            switch mode {
            case modeJson:
                if err := rest.loadParams(); err != nil {
                    l.Info("load params failed:%s", err.Error())
                    rest.SayError(http.StatusInternalServerError, "load params failed.")
                    return
                }
                if err := rest.decodeJson(); err != nil {
                    rest.SayError(http.StatusInternalServerError, "decode failed.")
                    return
                }
                if schema != nil {
                    if err := rest.authSchemaByte(schema); err != nil {
                        l.Info("auth request schema failed:%s", err.Error())
                        rest.SayError(http.StatusInternalServerError, "auth schema failed.")
                        return
                    }
                }
            case modeForm:
                err := r.ParseForm()
                if err != nil {
                    l.Info("parse form failed:%s", err.Error())
                    rest.SayError(http.StatusInternalServerError, "parse form failed.")
                    return
                }
            default:
                panic("coder fault http server mode miss match.")
            }
            h.ServeRest(rest)
        },
    }
}

func MakeRoute(path string, h RestHandler) {
    var j *validate.Property
    //if h impliments Validatable interface.
    if v, ok := h.(Validatable); ok {
        j = v.GetValidateSchema()
    }
    http.Handle(path, getHttpHandler(modeJson, h, j))
}

func MakeRouteForm(path string, h RestHandler) {
    http.Handle(path, getHttpHandler(modeForm, h, nil))
}
