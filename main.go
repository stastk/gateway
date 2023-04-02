package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.0"
	ProtoMajor int    // e.g. 1
	ProtoMinor int    // e.g. 0

	// response headers
	Header http.Header
	// response body
	Body io.ReadCloser
	// request that was sent to obtain the response
	Request *http.Request
}

var db = make(map[string]string)

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	//Remapper test
	r.GET("/remap/:version/:content/:direction", func(c *gin.Context) {

		versionStr := c.Params.ByName("version")
		_, versionIntErr := strconv.Atoi(versionStr)
		if versionStr == "" || versionIntErr != nil {
			c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP01"})
			return
		}

		contentStr := c.Params.ByName("content")
		contentMatch, _ := regexp.MatchString("(\\A[a-zA-Z]+=*\\z)", contentStr)
		if contentStr == "" || contentMatch == false {
			c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP02"})
			return
		}

		directionStr := c.Params.ByName("direction")
		directionOptions := []string{"gibberish", "normal"}
		fmt.Println("Here:")
		fmt.Println(contains(directionOptions, directionStr))
		if directionStr == "" || !contains(directionOptions, directionStr) {
			c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP03: " + directionStr})
			return
		}

		//var remapper_url = "http://ne3a.ru/remapper/v2?t=NDksOTcsMTAwLDEyMiw5Nw==&d=gibberish"
		//http://localhost:8080/remap/2/NDksOTcsMTAwLDEyMiw5Nw==/gibberish

		var url = "http://ne3a.ru/remapper/"
		res, err := http.Post(url+"v"+versionStr+"?t="+contentStr+"&d="+directionStr, "text; charset=UTF-8", c.Request.Body)
		//func Post(url, contentType string, body io.Reader) (*Response, error)

		// check for response error
		if err != nil {
			log.Fatal(err)
		}

		// read all response body
		data, _ := ioutil.ReadAll(res.Body)

		// close response body
		res.Body.Close()

		// print `data` as a string
		//fmt.Printf("%s\n", data)
		c.String(http.StatusOK, string(data))
	})

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ANSWER": "PONG"})
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	/* example curl for /admin with basicauth header
	   Zm9vOmJhcg== is base64("foo:bar")

		curl -X POST \
	  	http://localhost:8080/admin \
	  	-H 'authorization: Basic Zm9vOmJhcg==' \
	  	-H 'content-type: application/json' \
	  	-d '{"value":"bar"}'
	*/
	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			db[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
