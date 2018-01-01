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
