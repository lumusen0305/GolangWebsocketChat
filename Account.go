package main

import (
	"database/sql"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Claims struct {
	Account string `json:"account"`
	Username    string `json:"username"`
	jwt.StandardClaims
}
type Account struct {
	Id  string `json:"id"`
	Account  string `json:"account"  binding:"required"`
	Password string `json:"password"  binding:"required"`
	Email string `json:"email"  binding:"required"`
	Username string `json:"username"  binding:"required"`
}
type Login struct {
	Account  string `json:"account"`
	Password string `json:"password"  binding:"required"`
	Email string `json:"email"`
}
var jwtSecret = []byte("secret")
func AuthRequired(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	token := strings.Split(auth, "Bearer ")[1]
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (i interface{}, err error) {
		return jwtSecret, nil
	})

	if err != nil {
		var message string
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				message = "token is malformed"
			} else if ve.Errors&jwt.ValidationErrorUnverifiable != 0 {
				message = "token could not be verified because of signing problems"
			} else if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				message = "signature validation failed"
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				message = "token is expired"
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				message = "token is not yet valid before sometime"
			} else {
				message = "can not handle this token"
			}
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": message,
		})
		c.Abort()
		return
	}

	if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			fmt.Println("account:", claims.Account)
			fmt.Println("username:", claims.Username)
			c.Set("account", claims.Account)
			c.Set("username", claims.Username)
			c.Set("statue", true)
			c.Next()

	} else {
		c.Abort()
		return
	}
}
func register(c *gin.Context)  {
	pool, err := sqlx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	body := new(Account)
	if err := c.ShouldBindJSON(body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data": false,
			"code": http.StatusBadRequest,
			"msg": err.Error(),
		})
		return
	}

	uid,_ :=uuid.NewV4()
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
		return
	}
	encodePW := string(hashPassword)
	insertToUsers := `INSERT INTO users (id, account, password,email,username) VALUES (?,?,?, ?, ?)`
	_,errsql := pool.Exec(insertToUsers, uid.String(), body.Account,encodePW,body.Email,body.Username)
	if errsql != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data": false,
			"code": http.StatusBadRequest,
			"msg": errsql.Error(),
		})
		return
	}	else {
		c.JSON(http.StatusAccepted, gin.H{
			"data": true,
			"code": 200,
			"msg":"註冊成功",
		})
		return
	}
}
func login(c *gin.Context)  {
	p := Account{}
	pool, err := sqlx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	body := new(Login)
	if err := c.ShouldBindJSON(body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data": false,
			"code": http.StatusBadRequest,
			"msg": err.Error(),
		})
		return
	}
	//selectToUsers := `select * from users  where account=? AND password=?`

	errsql := pool.Get(&p,"SELECT * FROM users where account=?",body.Account)
	if errsql == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{
			"data": false,
			"code": http.StatusBadRequest,
			"msg": "查無帳號",
		})
		return
	}

	if errsql != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data": false,
			"code": http.StatusBadRequest,
			"msg": errsql.Error(),
		})
		return
	}else {

		errAccount := bcrypt.CompareHashAndPassword([]byte(p.Password), []byte(body.Password))
		if errAccount != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"data": false,
				"code": http.StatusBadRequest,
				"msg": "Unauthorized",
			})
	} else {
			now := time.Now()
			jwtId := body.Account + strconv.FormatInt(now.Unix(), 10)
			username := p.Username
			claims := Claims{
				Account: body.Account,
				Username:    username,
				StandardClaims: jwt.StandardClaims{
					Audience:  body.Account,
					ExpiresAt: now.Add(20 * time.Hour).Unix(),
					Id:        jwtId,
					IssuedAt:  now.Unix(),
					Issuer:    "ginJWT",
					NotBefore: now.Add(10 * time.Second).Unix(),
					Subject:   body.Account,
				},
			}
			tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			token, err := tokenClaims.SignedString(jwtSecret)
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{
					"data": false,
					"code": http.StatusBadRequest,
					"msg": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"token": token,
			})
			return
		}
		return
	}

}