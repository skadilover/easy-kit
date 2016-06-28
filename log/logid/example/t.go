package main

import (
    "fmt"
    "os"
)

func main1() {
    var a [3]byte
    a[0] = 'a'
    a[1] = 'a'
    a[2] = 'a'
    fmt.Println(string(a[:]))
    fmt.Println(os.Hostname())
}
