package main

import (
	"fmt"
	"gee"
	"html/template"
	"log"
	"net/http"
	"time"
)

func onlyForV2() gee.HandlerFunc {
	return func(c *gee.Context) {
		// start timer
		t := time.Now()
		// if a server error occur
		// c.Fail(500, "Internal Server Error")
		log.Printf("[%d] %s in %v for group v2", c.StatuCode, c.Req.RequestURI, time.Since(t))
	}
}

type student struct {
	Name string
	Age  int
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%2d-%2d", year, month, day)
}

func main() {
	r := gee.New()
	r.Use(gee.Logger()) // global middleware

	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})

	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	stu1 := &student{Name: "Geektutu", Age: 22}
	stu2 := &student{Name: "CHEN zi", Age: 21}
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})

	r.GET("/students", func(c *gee.Context) {
		c.HTML(http.StatusOK, "arr.tmpl", gee.H{
			"title":  "gee",
			"stuArr": []*student{stu1, stu2},
		})
	})

	r.GET("/date", func(c *gee.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", gee.H{
			"title": "what's up",
			"now":   time.Date(2001, 4, 9, 0, 0, 0, 0, time.UTC),
		})
	})

	v2 := r.Group("/v2")
	v2.Use(onlyForV2()) // v2 group middleware

	{
		v2.GET("/hello/:name", func(c *gee.Context) {
			// expect /hello/geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})

		v2.POST("/login", func(c *gee.Context) { // ???????????这个有问题
			c.JSON(http.StatusOK, gee.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})

	}

	r.Run(":9999")
}
