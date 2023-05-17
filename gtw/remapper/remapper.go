package remapper

import (
	"encoding/json"
	"fmt"
	contains "gateway/helpers"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// struct for #ver.1 and #ver.2 api
type RemapperResp struct {
	Text          []int
	DirectionTo   string
	DirectionFrom string
	Version       int
}

// route for sending data to us
var RemapperPath = "/remap/:version/:content/:direction_from"

// options of api versions we have
var versionOptions = []int{1, 2}
var versionInt int
var versionIntErr error
var directionOptions = []string{"gibberish", "normal"}

func GwRemap(c *gin.Context) {

	versionStr := c.Params.ByName("version")
	versionInt, versionIntErr = strconv.Atoi(versionStr)
	if versionStr == "" || versionIntErr != nil || !contains.ContainsInt(versionOptions, versionInt) {
		c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP01 in"})
		return
	}

	contentStr := c.Params.ByName("content")
	contentMatch, _ := regexp.MatchString("(\\A[a-zA-Z0-9]*=*\\z)", contentStr)
	if contentMatch == false {
		c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP02 in"})
		return
	}

	// `direction from`, because we don't know `direction to` when send request
	directionFromStr := c.Params.ByName("direction_from")
	if directionFromStr == "" || !contains.ContainsStr(directionOptions, strings.ToLower(directionFromStr)) {
		c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP03 in"})
		return
	}

	var url = "http://ne3a.ru/remapper/"
	var path = url + "v" + versionStr + "?t=" + contentStr + "&d=" + strings.ToLower(directionFromStr)
	var contentType = "text; charset=UTF-8"
	res, err := http.Post(path, contentType, c.Request.Body)

	// check for response error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"gw_err": "Error #RMP00-01"})
		log.Fatal(err)
	}

	// read all response body
	data, _ := io.ReadAll(res.Body)

	// close response body
	res.Body.Close()

	// --------------------//
	// Work with answer => //
	// --------------------//

	// try to parse response
	responseData := string(data)

	toMap := []byte(responseData)
	var mappedString map[string]interface{}
	if err := json.Unmarshal(toMap, &mappedString); err != nil {
		panic(err)
	}

	var textContent []int
	var directionTo string
	var directionFrom string

	// according to #ver. parse response in a different ways
	if versionInt == 1 {
		directionTo = fmt.Sprintf("%v", mappedString["direction"])
		directionFrom = fmt.Sprintf("%v", mappedString["invert_direction"])
		for i, value := range mappedString["text"].([]interface{}) {
			textContent = append(textContent, int(value.(float64)))
			i++
		}
	} else if versionInt == 2 {
		directionTo = fmt.Sprintf("%v", mappedString["direction_to"])
		directionFrom = fmt.Sprintf("%v", mappedString["direction_from"])
		for i, value := range mappedString["text"].([]interface{}) {
			textContent = append(textContent, int(value.(float64)))
			i++
		}
	}

	remapperResp := RemapperResp{
		Text:          textContent,
		DirectionTo:   directionTo,
		DirectionFrom: directionFrom,
		Version:       versionInt,
	}

	// validate directions
	if !contains.ContainsStr(directionOptions, strings.ToLower(remapperResp.DirectionTo)) || !contains.ContainsStr(directionOptions, strings.ToLower(remapperResp.DirectionFrom)) {
		c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP02 out"})
		return
	}

	// final request to remapper
	c.JSON(http.StatusOK, LastAnswer(remapperResp.Text, remapperResp.DirectionTo, remapperResp.DirectionFrom, remapperResp.Version))

}

// check #ver. before sending final request to remapper
func LastAnswer(text []int, to string, from string, version int) gin.H {
	var toBeSend gin.H
	if versionInt == 1 {
		toBeSend = gin.H{
			"text":             text,
			"direction":        strings.ToLower(to),
			"invert_direction": strings.ToLower(from),
			"version":          version,
		}
	} else if versionInt == 2 {
		toBeSend = gin.H{
			"text":           text,
			"direction_to":   strings.ToLower(to),
			"direction_from": strings.ToLower(from),
			"version":        version,
		}
	}

	return toBeSend
}
