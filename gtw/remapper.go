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

// Struct for #ver.1 and #ver.2 api
type RemapperResp struct {
	DirectionFrom string
	DirectionTo   string
	Text          []int
	Version       int
}

// Config: //

// Route for sending data to us
var RemapperPath = "/remap/:version/:content/:direction"

// Options of api versions we have
var versionOptions = []int{1, 2}

func SetPath(c *gin.Context) {

	versionStr := c.Params.ByName("version")
	versionInt, versionIntErr := strconv.Atoi(versionStr)
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
	//func Post(url, contentType string, body io.Reader) (*Response, error)

	// check for response error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"gw_err": "Error #RMP00-01"})
		log.Fatal(err)
	}

	// read all response body
	data, _ := io.ReadAll(res.Body)

	// close response body
	res.Body.Close()

	// try to parse response and send answer
	responseData := string(data)
	fmt.Println("HERE 002>")
	fmt.Println(responseData)
	var remapperRespErr error

	toMap := []byte(responseData)
	var mappedString map[string]interface{}
	if err := json.Unmarshal(toMap, &mappedString); err != nil {
		panic(err)
	}

	// According to #ver. parse response in a different ways
	var dirRr string
	var indirRr string
	var textRr []int
	if versionInt == 1 {
		dirRr = fmt.Sprintf("%v", mappedString["direction"])
		indirRr = fmt.Sprintf("%v", mappedString["invert_direction"])
		for i, value := range mappedString["text"].([]interface{}) {
			textRr = append(textRr, int(value.(float64)))
			i++
		}
	} else if versionInt == 2 {
		dirRr = fmt.Sprintf("%v", mappedString["direction_from"])
		indirRr = fmt.Sprintf("%v", mappedString["direction_to"])
		for i, value := range mappedString["text"].([]interface{}) {
			textRr = append(textRr, int(value.(float64)))
			i++
		}
	}

	remapperResp := RemapperResp{
		DirectionFrom: dirRr,
		DirectionTo:   indirRr,
		Text:          textRr,
	}

	if remapperRespErr != nil {
		c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP01 out"})
		return
	}
	if !contains.ContainsStr(directionOptions, strings.ToLower(remapperResp.DirectionTo)) {
		c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #RMP02 out"})
		return
	}

	// Check #ver. and fire final request
	if versionInt == 1 {
		c.JSON(http.StatusOK, gin.H{"text": remapperResp.Text, "direction": strings.ToLower(remapperResp.DirectionFrom), "invert_direction": strings.ToLower(remapperResp.DirectionTo)})
	} else if versionInt == 2 {
		c.JSON(http.StatusOK, gin.H{"text": remapperResp.Text, "direction_to": strings.ToLower(remapperResp.DirectionFrom), "direction_from": strings.ToLower(remapperResp.DirectionTo)})
	}

}
