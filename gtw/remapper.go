package remapper

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type RemapperResp struct {
	Direction       string `json:"direction"`        //To v1
	InvertDirection string `json:"invert_direction"` //From v1
	DirectionFrom   string `json:"direction_from"`   //To v2
	DirectionTo     string `json:"direction_to"`     //From v2
	Text            []int  `json:"text"`             //Text all versions
}

type RemapperRespV1 struct {
	Direction       string `json:"direction"`
	InvertDirection string `json:"invert_direction"`
	Text            []int  `json:"text"`
}

type RemapperRespV2 struct {
	DirectionFrom string `json:"direction_from"`
	DirectionTo   string `json:"direction_to"`
	Text          []int  `json:"text"`
}

var RemapperPath = "/remap/:version/:content/:direction"

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

func Remapper() *gin.Engine {
	r := gin.Default()
	//Remapper test

	r.POST(RemapperPath, func(c *gin.Context) {

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
}
