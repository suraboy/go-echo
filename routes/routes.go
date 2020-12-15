package routes

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/labstack/echo"
	"github.com/suraboy/go-echo/models"
	"github.com/suraboy/go-echo/config/mysql"
	"golang.org/x/crypto/bcrypt"
	"net/http"

)

func UserRoute(e *echo.Echo) {
	e.GET("/v1/users", GetAllUser)
	e.GET("/v1/users/:id", FindUser)
	e.POST("/v1/users", CreateUser)
	e.PUT("/v1/users/:id", UpdateUser)
	e.DELETE("/v1/users/:id", DeleteUser)
}

var messageError struct {
	Errors messageFormat `json:"errors"`
}

type messageFormat struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Error      string `json:"error"`
}

//get user list
func GetAllUser(c echo.Context) (err error) {
	db := mysql.connectDB()
	var user []models.Users
	db.Find(&user)
	return c.JSON(http.StatusOK, echo.Map{"datas": user})
}

//find uset by id
func FindUser(c echo.Context) (err error) {
	db := mysql.connectDB()
	id := c.Param("id")
	user := models.Users{}

	if err := db.Find(&user, id).Error; err != nil || db.Find(&user, id).RowsAffected == 0 {
		var msgError messageFormat
		if db.Find(&user, id).RowsAffected == 0 {
			msgError.StatusCode = http.StatusNotFound
			msgError.Message = "Not Found"
		} else {
			msgError.StatusCode = http.StatusInternalServerError
			msgError.Message = "Internal Server Error"
			msgError.Error = err.Error()
		}
		messageError.Errors = msgError
		return c.JSON(msgError.StatusCode, messageError)
	}

	return c.JSON(http.StatusOK, echo.Map{"data": user})
}

//create user
func CreateUser(c echo.Context) (err error) {
	db := mysql.connectDB()
	user := new(models.Users)
	if err = c.Bind(user); err != nil {
		var msgError messageFormat
		msgError.StatusCode = http.StatusBadRequest
		msgError.Message = "Bad Request"
		msgError.Error = err.Error()
		messageError.Errors = msgError
		return c.JSON(http.StatusBadRequest, messageError)
	}
	password := []byte(user.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	user.Password = string(hashedPassword)
	if err := db.Create(&user).Error; err != nil {
		var msgError messageFormat
		msgError.StatusCode = http.StatusExpectationFailed
		msgError.Message = "Expectation Failed"
		msgError.Error = err.Error()
		messageError.Errors = msgError
		return c.JSON(http.StatusExpectationFailed, messageError)
	} // pass pointer of data to Create

	return c.JSON(http.StatusCreated, echo.Map{"data": user})
}

//update user
func UpdateUser(c echo.Context) (err error) {
	pass := ""
	db := mysql.connectDB()
	id := c.Param("id")
	user := models.Users{}
	var msgError messageFormat

	if err := c.Bind(&user); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if user.Password != "" {
		password := []byte(user.Password)
		hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}
		pass = string(hashedPassword)
	}

	if err := db.Find(&user, id).Error; err != nil {
		msgError.StatusCode = http.StatusNotFound
		msgError.Message = "Not Found"
		messageError.Errors = msgError
		return c.JSON(msgError.StatusCode, messageError)
	}

	if pass != "" {
		user.Password = pass
	}

	if err := db.Save(&user).Error; err != nil {
		msgError.StatusCode = http.StatusExpectationFailed
		msgError.Message = "Expectation Failed"
		msgError.Error = err.Error()
		messageError.Errors = msgError
		return c.JSON(http.StatusExpectationFailed, messageError)
	}
	return c.JSON(http.StatusOK, echo.Map{"data": user})
}

//delete user
func DeleteUser(c echo.Context) (err error) {
	id := c.Param("id")
	db := mysql.connectDB()
	user := models.Users{}
	var msgError messageFormat
	if err := db.Find(&user, id).Error; err != nil {
		msgError.StatusCode = http.StatusNotFound
		msgError.Message = "Not Found"
		messageError.Errors = msgError
		return c.JSON(msgError.StatusCode, messageError)
	}

	if err := db.Delete(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	return c.NoContent(http.StatusNoContent)
}
