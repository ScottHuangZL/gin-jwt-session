# gin-jwt-session

[![GoDoc](https://godoc.org/github.com/ScottHuangZL/gin-jwt-session?status.png)](http://godoc.org/github.com/ScottHuangZL/gin-jwt-session)
[![Go Report](https://goreportcard.com/badge/github.com/ScottHuangZL/gin-jwt-session)](https://goreportcard.com/report/github.com/ScottHuangZL/gin-jwt-session)

## Description

gin-jwt-session is a Go package that provides JWT and Session for Gin framework

It already been adopt as one of gin [gin-contrib](https://github.com/gin-gonic/contrib) formal middleware


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

As for example please refer to [gin-jwt-session-example](https://github.com/ScottHuangZL/gin-jwt-session/tree/master/example). 

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

```go
package controllers

import (
	"github.com/ScottHuangZL/gin-jwt-session"
	"github.com/ScottHuangZL/gin-jwt-session/example/models"
	"github.com/gin-gonic/gin"
	// "log"
	"net/http"
	"time"
)

//LoginHandler for login page , it also can use for logout since it delete all stored session
func LoginHandler(c *gin.Context) {
	flashes := session.GetFlashes(c)
	session.DeleteAllSession(c)
	c.HTML(http.StatusOK, "home/login.html", gin.H{
		"title":   "Jwt Login",
		"flashes": flashes,
	})

}

//HomeHandler is the home handler
//will show home page, also according login/logout action to navigate
func HomeHandler(c *gin.Context) {
	// action := strings.ToLower(c.Param("action"))
	// path := strings.ToLower(c.Request.URL.Path)

	flashes := session.GetFlashes(c)
	username, err := session.ValidateJWTToken(c)
	loginFlag := false
	if err == nil && username != "" {
		loginFlag = true
	}
	c.HTML(http.StatusOK, "home/index.html", gin.H{
		"title":     "Main website",
		"now":       time.Now(),
		"flashes":   flashes,
		"loginFlag": loginFlag,
		"username":  username,
	})
}

//ValidateJwtLoginHandler validate the login and redirect to correct link
func ValidateJwtLoginHandler(c *gin.Context) {
	var form model.Login
	//try get login info
	if err := c.ShouldBind(&form); err != nil {
		session.SetFlash(c, "Get login info error: "+err.Error())
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}
	//validate login info
	if ok := model.ValidateUser(form.Username, form.Password); !ok {
		session.SetFlash(c, "Error : username or password")
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}
	//login info is correct, can generate JWT token and store in clien side now
	tokenString, err := session.GenerateJWTToken(form.Username, time.Hour*time.Duration(1))
	if err != nil {
		session.SetFlash(c, "Error Generate token string: "+err.Error())
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}

	err = session.SetTokenString(c, tokenString, 60*60) //60 minutes
	if err != nil {
		session.SetFlash(c, "Error set token string: "+err.Error())
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}
	session.SetFlash(c, "success : successful login")
	session.SetFlash(c, "username : "+form.Username)
	c.Redirect(http.StatusMovedPermanently, "/")
	return
}

//SomeCookiesHandler show cookie example
func SomeCookiesHandler(c *gin.Context) {
	session.Set(c, "hello", "world")
	sessionMessage, _ := session.GetString(c, "hello")
	session.Set(c, "hello", 2017)
	message2, _ := session.GetInt(c, "hello")
	session.Delete(c, "hello")
	readAgain, _ := session.GetString(c, "hello")
	c.JSON(http.StatusOK, gin.H{
		"session message":                 sessionMessage,
		"session new message":             message2,
		"session read again after delete": readAgain,
		"status": http.StatusOK})
}

```

## Credits

gin-jwt-session backend leverage 
[github.com/gorilla/sessions](https://github.com/gorilla/sessions)
[github.com/dgrijalva/jwt-go](https://github.com/gorilla/sessions)