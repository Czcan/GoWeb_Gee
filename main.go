package main

import (
	"fmt"
	"gee" // 导入私有包
	"net/http"
)

func main() {
	r := gee.New()

	r.GET("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "r.URL.Path:  %q\n", r.URL.Path)
	})

	r.POST("/hello", func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	})

	r.Run(":8088")
}
