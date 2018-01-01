package model

//Login is a simple type to contain username and password
type Login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

//ValidateUser validate the username and password is valid or not
func ValidateUser(username, password string) bool {
	if username == "admin" && password == "admin" ||
		username == "user1" && password == "user1" ||
		username == "user2" && password == "user2" {
		return true
	}
	return false
}
