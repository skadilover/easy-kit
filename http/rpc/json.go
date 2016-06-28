package rpc

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/bitly/go-simplejson"
    "github.com/skadilover/easykit/log"
    "github.com/skadilover/easykit/validate"
)

var restClient = &http.Client{
    Timeout: time.Second * 10,
    Transport: &http.Transport{
        MaxIdleConnsPerHost:   200,
        ResponseHeaderTimeout: time.Second * 10,
        DisableKeepAlives:     false,
    },
}

func BasicHttpGet(rawurl string, l log.Logger) ([]byte, error) {
    u, err := url.Parse(rawurl)
    tmp := u.Query()
    u.RawQuery = tmp.Encode()
    req := newHttpRequest("GET", u, nil)
    l.Info("request [%v]", u)
    t1 := time.Now()
    response, err := restClient.Do(req)
    if err != nil {
        l.Error("Post [%s] failed:%s", rawurl, err.Error())
        return nil, err
    }
    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        l.Error("read form [%s] response failed:%s", rawurl, err.Error())
        return nil, err
    }
    if response.StatusCode != http.StatusOK {
        l.Error("Post [%s] failed: http code %s", rawurl, responseData)
        return nil, fmt.Errorf("http status is not ok.")
    }
    t2 := time.Now()
    sub := t2.Sub(t1).Nanoseconds() / 1000000
    l.Info("[%s] response is :%s time cost is %v ms", rawurl, responseData, sub)
    return responseData, nil
}

func TextHttpPost(url, text string, l log.Logger) ([]byte, error) {
    data := []byte(text)
    response, err := http.Post(url, "application/json", bytes.NewBuffer(data))
    if err != nil {
        l.Error("Post [%s] failed:%s", url, err.Error())
        return nil, err
    }
    l.Info("%s request is %s", url, data)
    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        l.Error("read form [%s] response failed:%s", url, err.Error())
        return nil, err
    }
    l.Info("[%s] response is :%s", url, responseData)
    return responseData, nil

}

func newHttpRequest(method string, u *url.URL, body io.Reader) *http.Request {
    switch method {
    case "POST":
        rc, ok := body.(io.ReadCloser)
        if !ok && body != nil {
            rc = ioutil.NopCloser(body)
        }
        req := &http.Request{
            Method:     method,
            URL:        u,
            Proto:      "HTTP/1.1",
            ProtoMajor: 1,
            ProtoMinor: 1,
            Header:     make(http.Header),
            Body:       rc,
            Host:       u.Host,
        }
        req.Header.Set("Content-Type", "application/json")
        if body != nil {
            switch v := body.(type) {
            case *bytes.Buffer:
                req.ContentLength = int64(v.Len())
            case *bytes.Reader:
                req.ContentLength = int64(v.Len())
            case *strings.Reader:
                req.ContentLength = int64(v.Len())
            }
        }
        return req
    case "GET":
        req := &http.Request{
            Method:     method,
            URL:        u,
            Proto:      "HTTP/1.1",
            ProtoMajor: 1,
            Header:     make(http.Header),
            ProtoMinor: 1,
            Host:       u.Host,
        }
        return req
    default:
        return nil
    }
}

func JsonHttpPost(rawurl string, m interface{}, l log.Logger) ([]byte, error) {
    data, err := json.Marshal(m)
    if err != nil {
        l.Error("marshal [%s] request failed:%s", rawurl, err.Error())
        return nil, err
    }
    u, err := url.Parse(rawurl)
    tmp := u.Query()
    tmp.Set("logid", l.Logid())
    u.RawQuery = tmp.Encode()
    req := newHttpRequest("POST", u, bytes.NewBuffer(data))
    l.Info("[%v] %s", u, data)
    t1 := time.Now()
    response, err := restClient.Do(req)
    if err != nil {
        l.Error("Post [%s] failed:%s", rawurl, err.Error())
        return nil, err
    }
    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        l.Error("read form [%s] response failed:%s", rawurl, err.Error())
        return nil, err
    }
    if response.StatusCode != http.StatusOK {
        l.Error("Post [%s] failed: http code %s", rawurl, responseData)
        return nil, fmt.Errorf("http status is not ok.")
    }
    t2 := time.Now()
    sub := t2.Sub(t1).Nanoseconds() / 1000000
    l.Info("[%s] response is :%s time cost is %v ms", rawurl, responseData, sub)
    return responseData, nil
}

//return simplejson object
func SimpleJsonHttpPost(url string, request interface{}, l log.Logger) (*simplejson.Json, error) {
    data, err := JsonHttpPost(url, request, l)
    if err != nil {
        return nil, err
    }
    j, err := simplejson.NewJson(data)
    if err != nil {
        return nil, fmt.Errorf("decode json failed:%s", err.Error())
    }
    return j, nil
}

//validate response.
func JsonPostValidate(url string, request interface{}, p *validate.Property, l log.Logger) (*simplejson.Json, error) {
    data, err := JsonHttpPost(url, request, l)
    if err != nil {
        return nil, err
    }
    if p != nil {
        if err := p.ValidateString(string(data)); err != nil {
            l.Error("validate failed:%s", err.Error())
            return nil, err
        }
    }
    j, err := simplejson.NewJson(data)
    if err != nil {
        return nil, fmt.Errorf("decode json failed:%s", err.Error())
    }
    return j, nil

}
