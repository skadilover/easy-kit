package main

import (
    "easy-kit/log"
    "errors"
    "time"
)

func main() {
    h := log.LogHeader{
        LogId:  "123456",
        ReqId:  "123.18.10.198",
        HostId: "127.0.0.1",
    }
    log.Error(h, "the count is %d", 123)
    log.Info(h, "the info is [%s][%d][%v]", "xxx", 123, errors.New("xxxx"))
    time.Sleep(3 * time.Second)
}
