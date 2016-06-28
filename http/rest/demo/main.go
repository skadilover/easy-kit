package main

import (
    "fmt"
    "net/http"

    "github.com/skadilover/easykit/http/rest"
    "github.com/skadilover/easykit/log"
    "github.com/skadilover/easykit/validate"
)

//HelloWorldHandler
type HelloWorldHandler struct{}

//RestHandler interface.
func (h *HelloWorldHandler) ServeRest(r rest.Rest) error {
    //读取请求字段
    msg := r.JsonInput().Get("message").MustString("defualt message.")
    name := r.JsonInput().GetPath("payload", "name").MustString("no name.")
    //写日志
    r.Info("recieving messeage:%s name :%s", msg, name)
    response := map[string]interface{}{
        "status":  200,
        "message": "you call me.",
    }
    //写应答,该函数已自动日志了返回的报文
    r.SayJson(response)
    return nil
}

//Validatable interface.
func (h *HelloWorldHandler) GetValidateSchema() *validate.Property {
    return schema
}

var schema = validate.NewProperty(validate.TypeObject).Properties(
    map[string]*validate.Property{
        "name": validate.NewProperty(validate.TypeString).MaxLength(3),
        "age":  validate.NewProperty(validate.TypeNumber).Max(99),
    }).Required("name", "age")

func main() {
    fmt.Println("begin...")
    //初始化日志
    log.Initialize_Base_Logger("./", "test_server.json", 1, log.DEBUG)
    //创建映射
    rest.MakeRoute("/test/hello", &HelloWorldHandler{})
    ch := make(chan error)
    go func() {
        err := http.ListenAndServe(fmt.Sprintf(":%d", 47897), nil)
        if err != nil {
            ch <- err
        }
    }()
    err := <-ch
    fmt.Println("order server shutting down,error:", err)
}
