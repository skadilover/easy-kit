package log

import (
    "net/http"
    "strings"
    "testing"

    "github.com/skadilover/easykit/log/logid"
)

type Logger interface {
    Error(format string, v ...interface{})
    Info(format string, v ...interface{})
    Tag(tag string, msg interface{})
    Logid() string
    Head() LogHeader
}

func getRealIp(r *http.Request) string {
    remoteAddr := ""
    forward := r.Header.Get("X-Forwarded-For")
    ips := strings.Split(forward, ",")
    if len(ips) > 0 {
        if len(ips[0]) > 0 {
            remoteAddr = ips[0]
        }
    }
    return remoteAddr
}

func GetHttpLogger(r *http.Request) Logger {
    logId := r.URL.Query().Get("logid")
    if logId == "" {
        logId = logid.NewObjectId().Hex()
    }
    cid := r.URL.Query().Get("cid")
    module := r.URL.Path
    return &httpLogger{
        h: LogHeader{
            LogId:    logId + "###" + cid,
            Module:   module,
            Lat:      r.URL.Query().Get("lat"),
            Lng:      r.URL.Query().Get("lng"),
            CallerIp: getRealIp(r),
        },
    }
}

type httpLogger struct {
    h LogHeader
}

func (l *httpLogger) Logid() string {
    return l.h.LogId
}

func (l *httpLogger) Head() LogHeader {
    return l.h
}

func (l *httpLogger) Error(format string, v ...interface{}) {
    Error(l.h, format, v...)

}

func (l *httpLogger) Info(format string, v ...interface{}) {
    Info(l.h, format, v...)
}

func (l *httpLogger) Tag(tag string, msg interface{}) {
    Tag(l.h, tag, msg)
}

func GetTestLogger(t *testing.T) Logger {
    return &testLogger{t: t}
}

func (l *testLogger) Logid() string {
    return "test"
}

func (l *testLogger) Head() LogHeader {
    return LogHeader{}
}

type testLogger struct {
    t *testing.T
}

func (l *testLogger) Error(format string, v ...interface{}) {
    l.t.Errorf(format, v...)

}

func (l *testLogger) Info(format string, v ...interface{}) {
    l.t.Logf(format, v...)
}

func (l *testLogger) Tag(tag string, msg interface{}) {
    l.t.Log(tag, msg)
}
