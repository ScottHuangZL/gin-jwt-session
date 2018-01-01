# gin-jwt-session

[![GoDoc](https://godoc.org/github.com/ScottHuangZL/gin-jwt-session?status.png)](http://godoc.org/github.com/ScottHuangZL/gin-jwt-session)


## Description

gin-jwt-session is a Go package that provides JWT and Session for Gin framework
It has the following features:

* Provide JWT tokenstring generate and validate function.
* Provide defautlt secured session
* Provide secured flashes functions
* Simple to use and also allow you manual adjust options if necessary
* Provide sample for you to easily follow up

## Requirements

Go 1.6 or above.

## Installation

Run the following command to install the package:

```
go get github.com/ScottHuangZL/gin-jwt-session
```

## Getting Started

As for example please refer to [gin-jwt-session-example](https://github.com/ScottHuangZL/gin-jwt-session/example). 

```go
package main

import (
	"context"
	"github.com/ScottHuangZL/gin-jwt-session"
	"github.com/ScottHuangZL/gin-jwt-session/example/controllers"
	"github.com/ScottHuangZL/gin-jwt-session/example/models"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	r := gin.Default()
	//below are optional setting, you change it or just comment it to let it as default
	// session.SecretKey = "You any very secriet key !@#$!@%@"  //Any characters
	// session.JwtTokenName = "YouCanChangeTokenName"               //no blank character
	// session.DefaultFlashSessionName = "YouCanChangeTheFlashName" //no blank character
	// session.DefaultSessionName = "YouCanChangeTheSessionName"    //no blank character
	//end of optional setting
	session.NewStore()
	r.Use(session.ClearMiddleware()) //important to avoid mem leak
	setupRouter(r)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		// service connections
		if err := s.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func setupRouter(r *gin.Engine) {
	r.Delims("{%", "%}")
	// Default With the Logger and Recovery middleware already attached
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	r.MaxMultipartMemory = 8 << 20 // 8 MiB
	r.Static("/static", "./static")
	r.SetFuncMap(template.FuncMap{
		"formatAsDate": model.FormatAsDate,
	})
	r.LoadHTMLGlob("views/**/*")
	r.GET("/login", controllers.LoginHandler)
	r.GET("/logout", controllers.LoginHandler) //logout also leverage login handler, since it just need clear session
	r.POST("/validate-jwt-login", controllers.ValidateJwtLoginHandler)
	r.GET("/index.html", controllers.HomeHandler)
	r.GET("/index", controllers.HomeHandler)
	r.GET("", controllers.HomeHandler)

	r.GET("/some-cookie-example", controllers.SomeCookiesHandler)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ping": "pong"})
	})

}
```

## Credits

gin-jwt-session backend leverage 
[github.com/gorilla/sessions](https://github.com/gorilla/sessions)
[github.com/dgrijalva/jwt-go](https://github.com/gorilla/sessions)