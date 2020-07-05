package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/jmoiron/sqlx"
	"net/http"
	"strings"
)
var dsn string= "root:0305@tcp(localhost:3306)/golang?charset=utf8&parseTime=True&loc=Local"
var pool *sqlx.DB
var (
	router = gin.New()
)
func main() {
	go hub.run()
	router.Use(Cors())
	account := router.Group("/account")
	account.POST("/register", func(c *gin.Context) {
		register(c)
	})
	account.POST("/login", func(c *gin.Context) {
		login(c)
	})
	router.GET("/ws/:Sender", func(c *gin.Context) {
		Sender := c.Param("Sender")
		serveWs(c.Writer, c.Request,Sender)
	})
	// protected member router
	account.Use(AuthRequired)
	{
		account.POST("/member/profile", func(c *gin.Context) {
			if c.MustGet("statue") ==true {
				c.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"data": gin.H{
						"account":  c.MustGet("account"),
						"username": c.MustGet("username"),
					},
					"msg": "登入成功",

				})
				return
			}
			c.JSON(http.StatusNotFound, gin.H{
				"error": "can not find the record",
			})
		})
	}
	router.Use(cors.Default())


	router.Run("0.0.0.0:12345")
}
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method      //请求方法
		origin := c.Request.Header.Get("Origin")        //请求头部
		var headerKeys []string                             // 声明请求头keys
		for k, _ := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "*")        // 这是允许访问所有域
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")      //服务器支持的所有跨域请求的方法,为了避免浏览次请求的多次'预检'请求
			//  header的类型
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			//              允许跨域设置                                                                                                      可以返回其他子段
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar")      // 跨域关键设置 让浏览器可以解析
			c.Header("Access-Control-Max-Age", "172800")        // 缓存请求信息 单位为秒
			c.Header("Access-Control-Allow-Credentials", "false")       //  跨域请求是否需要带cookie信息 默认设置为true
			c.Set("content-type", "application/json")       // 设置返回格式是json
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		// 处理请求
		c.Next()        //  处理请求
	}
}