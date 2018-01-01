package session

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
	gorillaContext "github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"sync"
	"time"
)

const defaultShortName = "default"

var (
	once sync.Once
	//SecretKey to encrypt session
	//please clear the browser buffer(ctrl+shift+delete) in case you change the key and token name to new one
	SecretKey = "welcome to Scott Huang's Session and JWT util, please change to your secret accodingly."

	//JwtTokenName is JWT token name, also is the JWT session name too
	//Not too long and also do not include blank, such as do not set as "a blank name"
	JwtTokenName = "jwtTokenSession"

	//DefaultFlashSessionName to store the flash session
	//Not too long and also do not include blank, such as do not set as "a blank name"
	DefaultFlashSessionName = "myDefaultFlashSessionName"

	//DefaultSessionName to store other session message
	//Not too long and also do not include blank, such as do not set as "a blank name"
	DefaultSessionName = "myDefaultSessionName"

	//DefaultOption to provide default option
	//Maxage set the session duration by second
	DefaultOption = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 1, //1 hour
		HttpOnly: true,
	}

	//store for the app sessions
	store *sessions.CookieStore
)

func init() {
	//TODO
}

//ClearMiddleware clear mem to avoid leak.
//you should add this middleware at your main gin.router
//
// Please see note from http://www.gorillatoolkit.org/pkg/sessions
// Important Note: If you aren't using gorilla/mux,
// you need to wrap your handlers with context.ClearHandler as or else you will leak memory!
// ClearHandler actually invoke Clear func
func ClearMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		defer gorillaContext.Clear(c.Request)
		c.Next()
	}
}

//NewStore only create one session store globally
//The main program need call sessions.NewStore() to initialize store
//User also can set sessions var before call this function to adjust some default parameters
//
//Usage Sample:
// 	//in your main(), setup session after r = gin.Default()
// 	sessions.JwtTokenName = "YourJWTTokenName"                       //string without blank
// 	sessions.DefaultSessionName = "YourDefaultSessionName"           //string without blank
// 	sessions.DefaultFlashSessionName = "YourDefaultFlashSessionName" //string without blank
// 	sessions.SecretKey = "Your Secerect Key (*&%(&*%$"               //string with any
// 	sessions.NewStore()                                              //setup the session store
// 	r.Use(sessions.ClearMiddleware())                                //important to avoid memory leak
// 	//end setup session
func NewStore() {
	once.Do(newStore)
}

//internal only
func newStore() {
	store = sessions.NewCookieStore([]byte(SecretKey))
}

//Message struct contain message you would like to set in session
//usually you just provide Key and Value only
type Message struct {
	Key         interface{}
	Value       interface{}
	SessionName string
	Options     *sessions.Options
}

//Flash type to contain new flash and its session
type Flash struct {
	Flash       interface{}
	SessionName string
}

// Options stores configuration for a session or session store.
// Fields are a subset of http.Cookie fields.
type Options = sessions.Options

//SetMessage will set session into gin.Context per gived SesionMessage
//It is the basic function for other set func to leverage
//
//Usage Sample:
//  err := sessions.SetMessage(c,
// 	sessions.Message{
// 		Key:   sessions.JwtTokenName,
// 		Value: tokenString,
// 		// SessionName: "",
// 		// Options: &sessions.Options{
// 		// 	Path:     "/",
// 		// 	MaxAge:   3600 * 1, //1 hour for session. Btw, the token itself have valid period, so, better set it as same
// 		// 	HttpOnly: true,
// 		// },
// 	})
func SetMessage(c *gin.Context, message Message) (err error) {
	if len(message.SessionName) == 0 || message.SessionName == defaultShortName {
		message.SessionName = DefaultSessionName
	}
	session, err := store.Get(c.Request, message.SessionName)
	if err != nil {
		return
	}
	if message.Options == nil {
		session.Options = DefaultOption
	} else {
		session.Options = message.Options
	}
	session.Values[message.Key] = message.Value
	err = session.Save(c.Request, c.Writer)
	return err
}

//SetSessionFlash set new flash to givied session
func SetSessionFlash(c *gin.Context, flash Flash) (err error) {
	if len(flash.SessionName) == 0 || flash.SessionName == defaultShortName {
		flash.SessionName = DefaultFlashSessionName
	}
	session, err := store.Get(c.Request, flash.SessionName)
	if err != nil {
		return
	}
	session.AddFlash(flash.Flash, flash.SessionName)
	err = session.Save(c.Request, c.Writer)
	return err
}

//SetFlash set new flash to default session
func SetFlash(c *gin.Context, flash interface{}) (err error) {
	return SetSessionFlash(c, Flash{
		Flash:       flash,
		SessionName: DefaultFlashSessionName,
	})
}

//GetSessionFlashes return previously flashes per gived session
func GetSessionFlashes(c *gin.Context, sessionName string) []interface{} {
	if len(sessionName) == 0 || sessionName == defaultShortName {
		sessionName = DefaultFlashSessionName
	}
	session, err := store.Get(c.Request, sessionName)
	if err != nil {
		return nil
	}
	flashes := session.Flashes(sessionName)
	DeleteSession(c, sessionName) //manual delete flash session
	return flashes
}

//GetFlashes return previously flashes from default session
func GetFlashes(c *gin.Context) []interface{} {
	return GetSessionFlashes(c, DefaultFlashSessionName)
}

//DeleteSession delete a session
//instead of Set, the delete will set MaxAge to -1
func DeleteSession(c *gin.Context, sessionName string) (err error) {
	if len(sessionName) == 0 || sessionName == defaultShortName {
		sessionName = DefaultSessionName
	}
	session, err := store.Get(c.Request, sessionName)
	if err != nil {
		return
	}
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1, //it mean delete
		HttpOnly: true,
	}
	// Save it before we write to the response/return from the handler.
	err = session.Save(c.Request, c.Writer)
	return err
}

//GetSessionValue session value according
//The base get func, other will leverage this func
func GetSessionValue(c *gin.Context, sessionName string, key interface{}) (value interface{}, err error) {
	// Get a session. Get() always returns a session, even if empty.
	if len(sessionName) == 0 {
		sessionName = DefaultSessionName
	}
	session, err := store.Get(c.Request, sessionName)
	if err != nil {
		return nil, err
	}
	value = session.Values[key]
	return
}

//DeleteSessionValue try delete the gived session key value
func DeleteSessionValue(c *gin.Context, sessionName string, key interface{}) (err error) {
	if len(sessionName) == 0 {
		sessionName = DefaultSessionName
	}
	session, err := store.Get(c.Request, sessionName)
	if err != nil {
		return err
	}
	delete(session.Values, key)
	return nil
}

//Delete try delete default session key value
func Delete(c *gin.Context, key interface{}) (err error) {
	return DeleteSessionValue(c, DefaultSessionName, key)
}

//GetDefaultSessionValue levarage GetSessionValue to get key value
func GetDefaultSessionValue(c *gin.Context, key interface{}) (value interface{}, err error) {
	return GetSessionValue(c, DefaultSessionName, key)
}

//GetString levarage Get to get key value and convert to string
func GetString(c *gin.Context, key interface{}) (value string, err error) {
	valueInterface, err := GetDefaultSessionValue(c, key)
	if err != nil {
		return "", err
	}
	value, ok := valueInterface.(string)
	if !ok {
		err = errors.New("convert session value to string failed")
	}
	return value, err
}

//GetTokenString to get tokenString
func GetTokenString(c *gin.Context) (tokenString string, err error) {
	valueInterface, err := GetSessionValue(c, JwtTokenName, JwtTokenName)
	if err != nil {
		tokenString = ""
		return tokenString, err
	}
	tokenString, ok := valueInterface.(string)
	if !ok {
		err = errors.New("failed get token string")
	}
	return tokenString, err
}

//SetTokenString into JwtTokenSession
func SetTokenString(c *gin.Context, tokenString string, seconds int) (err error) {
	return SetMessage(c, Message{
		Key:         JwtTokenName,
		Value:       tokenString,
		SessionName: JwtTokenName,
		Options: &Options{
			Path:     "/",
			MaxAge:   seconds,
			HttpOnly: true,
		},
	})
}

//DeleteTokenSession delete before generate JWT token string
func DeleteTokenSession(c *gin.Context) (err error) {
	return DeleteSession(c, JwtTokenName)
}

//DeleteNormalSession will delete DefaultSessionName and DefaultFlashSessionName
func DeleteNormalSession(c *gin.Context) {
	DeleteSession(c, DefaultSessionName)
	DeleteSession(c, DefaultFlashSessionName)
}

//DeleteAllSession will delete JwtTokenName/DefaultSessionName/DefaultFlashSessionName
//usually used when user logout
func DeleteAllSession(c *gin.Context) {
	DeleteSession(c, JwtTokenName)
	DeleteNormalSession(c)
}

//GetInt levarage Get to get key value and convert to int
func GetInt(c *gin.Context, key interface{}) (value int, err error) {
	valueInterface, err := GetDefaultSessionValue(c, key)
	if err != nil {
		return 0, err
	}
	value, ok := valueInterface.(int)
	if !ok {
		err = errors.New("convert session value to string failed")
	}
	return value, err
}

//Set set key value by interface
func Set(c *gin.Context, key, value interface{}) (err error) {
	return SetMessage(c, Message{
		Key:   key,
		Value: value,
	})
}

//GenerateJWTToken per gived username and token duration
//
// Claims example as below, and we only use below top 3 claims
//   "iat": 1416797419, //start
//   "exp": 1448333419, //end
//   "sub": "jrocket@example.com",  //username
//
//   "iss": "Online JWT Builder",
//   "aud": "www.example.com",
//   "GivenName": "Johnny",
//   "Surname": "Rocket",
//   "Email": "jrocket@example.com",
//   "Role": [ "Manager", "Project Administrator" ]
//
func GenerateJWTToken(username string, tokenDuration time.Duration) (tokenString string, err error) {
	//start generate jwt token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(tokenDuration).Unix()
	claims["iat"] = time.Now().Unix()
	claims["sub"] = username
	//claims["email"] = form.Email
	token.Claims = claims
	tokenString, err = token.SignedString([]byte(SecretKey))
	return tokenString, err
}

//ValidateJWTToken valide JWT per headder firstly
//And then try get from session if above failed
//will return valid username in case no err
//username == "" also mean failed
func ValidateJWTToken(c *gin.Context) (username string, err error) {
	//get username firstly
	username = ""

	//then try valid token
	tokenString := c.Request.Header.Get("Authorization")
	if "" == tokenString {
		//Not in header authorization, try get cookie in case header empty
		tokenString, err = GetTokenString(c)
		if err != nil {
			// log.Println("Validate JWT Error: failed get token from session: ", err.Error())
			return username, err
		}
		// log.Println("success get tokenString: ", tokenString)
	}

	// log.Println("After get cookie")
	// Get token from request
	c.Request.Header.Set("Authorization", tokenString)
	token, err := request.ParseFromRequest(c.Request, request.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SecretKey), nil
		})
	if err != nil {
		return username, err
	}
	if !token.Valid {
		return username, errors.New("token not valid")
	}
	// log.Println("Token claims: ", token.Claims)
	err = errors.New("failed to fetch username from token")
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if _username, ok := claims["sub"].(string); ok {
			err = nil
			username = _username
		}
	}
	return username, err
}
