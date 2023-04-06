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
	DirectionFrom string
	DirectionTo   string
	Text          []int
	Version       int
}

// route for sending data to us
var RemapperPath = "/remap/:version/:content/:direction"

// options of api versions we have
var versionOptions = []int{1, 2}
var versionInt int
var versionIntErr error

func GwRemap(c *gin.Context) {

	versionStr := c.Params.ByName("version")
	if versionStr == "latest" || versionStr == "" {
		versionInt = versionOptions[len(versionOptions)-1]
		versionStr = strconv.Itoa(versionInt)
	} else {
		versionInt, versionIntErr = strconv.Atoi(versionStr)
	}

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

	directionStr := c.Params.ByName("direction")
	directionOptions := []string{"gibberish", "normal"}
	if directionStr == "" || !contains.ContainsStr(directionOptions, strings.ToLower(directionStr)) {
		c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP03 in"})
		return
	}

	var url = "http://ne3a.ru/remapper/"
	var path = url + "v" + versionStr + "?t=" + contentStr + "&d=" + strings.ToLower(directionStr)
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

	// try to parse response
	responseData := string(data)

	toMap := []byte(responseData)
	var mappedString map[string]interface{}
	if err := json.Unmarshal(toMap, &mappedString); err != nil {
		panic(err)
	}

	// according to #ver. parse response in a different ways
	var directionFrom string
	var directionTo string
	var textContent []int
	if versionInt == 1 {
		directionFrom = fmt.Sprintf("%v", mappedString["direction"])
		directionTo = fmt.Sprintf("%v", mappedString["invert_direction"])
		for i, value := range mappedString["text"].([]interface{}) {
			textContent = append(textContent, int(value.(float64)))
			i++
		}
	} else if versionInt == 2 {
		directionFrom = fmt.Sprintf("%v", mappedString["direction_from"])
		directionTo = fmt.Sprintf("%v", mappedString["direction_to"])
		for i, value := range mappedString["text"].([]interface{}) {
			textContent = append(textContent, int(value.(float64)))
			i++
		}
	}

	remapperResp := RemapperResp{
		DirectionFrom: directionFrom,
		DirectionTo:   directionTo,
		Text:          textContent,
		Version:       versionInt,
	}

	// validate directions
	if !contains.ContainsStr(directionOptions, strings.ToLower(remapperResp.DirectionTo)) || !contains.ContainsStr(directionOptions, strings.ToLower(remapperResp.DirectionFrom)) {
		c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP02 out"})
		return
	}

	// check #ver. and fire final request to remapper
	if versionInt == 1 {
		c.JSON(http.StatusOK, gin.H{
			"text":             remapperResp.Text,
			"direction":        strings.ToLower(remapperResp.DirectionTo),
			"invert_direction": strings.ToLower(remapperResp.DirectionFrom),
			"version":          remapperResp.Version,
		})
	} else if versionInt == 2 {
		c.JSON(http.StatusOK, gin.H{
			"text":           remapperResp.Text,
			"direction_to":   strings.ToLower(remapperResp.DirectionTo),
			"direction_from": strings.ToLower(remapperResp.DirectionFrom),
			"version":        remapperResp.Version,
		})
	}

}
