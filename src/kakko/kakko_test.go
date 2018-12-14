package main_test

import (
    "testing"
    "fmt"
    "time"
)

func TestCurd(t *testing.T) {
    fmt.Println("running testing Curd... ")
    //main.Curd()
    // 单位是秒
    fmt.Println(int64(time.Now().Unix()))
    fmt.Println(int64(time.Now().UnixNano() / 1e6))
    //fmt.Println(time.Time{})
}

func TestCalculate(t *testing.T) {
    fmt.Println("running testing Calculate... ")
}

func TestGetElasticsearch5M(t *testing.T) {
    fmt.Println("running testing TestGetElasticsearch5M... ")

}
