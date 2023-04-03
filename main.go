package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

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

type RemapperResp struct {
	Direction       string `json:"direction"`        //To v1
	InvertDirection string `json:"invert_direction"` //From v1
	DirectionFrom   string `json:"direction_from"`   //To v2
	DirectionTo     string `json:"direction_to"`     //From v2
	Text            []int  `json:"text"`             //Text all versions
}

var db = make(map[string]string)

func containsStr(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func containsInt(s []int, str int) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	//Remapper test
	r.POST("/remap/:version/:content/:direction", func(c *gin.Context) {

		versionStr := c.Params.ByName("version")
		versionInt, versionIntErr := strconv.Atoi(versionStr)
		versionOptions := []int{1, 2}
		if versionStr == "" || versionIntErr != nil || !containsInt(versionOptions, versionInt) {
			c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP01 in"})
			return
		}

		contentStr := c.Params.ByName("content")
		contentMatch, _ := regexp.MatchString("(\\A[a-zA-Z0-9]*=*\\z)", contentStr)
		if contentMatch == false {
			c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP02 in"})
			return
		}

		directionStr := c.Params.ByName("direction")
		directionOptions := []string{"gibberish", "normal"}
		if directionStr == "" || !containsStr(directionOptions, strings.ToLower(directionStr)) {
			c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP03 in"})
			return
		}

		//var remapper_url = "http://ne3a.ru/remapper/v2?t=NDksOTcsMTAwLDEyMiw5Nw==&d=gibberish"
		//http://localhost:8080/remap/2/NDksOTcsMTAwLDEyMiw5Nw==/gibberish

		var url = "http://ne3a.ru/remapper/"
		res, err := http.Post(url+"v"+versionStr+"?t="+contentStr+"&d="+strings.ToLower(directionStr), "text; charset=UTF-8", c.Request.Body)
		//func Post(url, contentType string, body io.Reader) (*Response, error)

		// check for response error
		if err != nil {
			log.Fatal(err)
		}

		// read all response body
		data, _ := ioutil.ReadAll(res.Body)

		// close response body
		res.Body.Close()

		responseData := string(data)
		//responseData := `{"direction":"1", "text":"657", "invert_direction":"something"}`

		var remapperResp RemapperResp

		remapperRespErr := json.Unmarshal([]byte(responseData), &remapperResp)
		if remapperRespErr != nil {
			c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP01 out"})
			return
		}
		if !containsStr(directionOptions, strings.ToLower(remapperResp.Direction)) && !containsStr(directionOptions, strings.ToLower(remapperResp.DirectionTo)) {
			c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP02 out"})
			return
		}
		var processedResponse gin.H
		if versionInt == 1 {
			processedResponse = gin.H{"text": remapperResp.Text, "direction": strings.ToLower(remapperResp.Direction), "invert_direction": strings.ToLower(remapperResp.InvertDirection)}
		} else if versionInt == 2 {
			processedResponse = gin.H{"text": remapperResp.Text, "direction_to": strings.ToLower(remapperResp.DirectionTo), "direction_from": strings.ToLower(remapperResp.DirectionFrom)}
		}

		c.JSON(http.StatusOK, processedResponse)
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
