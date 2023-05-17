package prania

import (
	contains "gateway/helpers"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// route for sending data to us
var PraniaPath = "/prania/:version/:entity/:action/:id"
var versionOptions = []int{1}
var versionInt int
var versionIntErr error

var entityOptions = []string{"list", "ingridient"}
var actionOptions = []string{"show"}

func GwGetEntity(c *gin.Context) {
	versionStr := c.Params.ByName("version")
	versionInt, versionIntErr = strconv.Atoi(versionStr)
	if versionStr == "" || versionIntErr != nil || !contains.ContainsInt(versionOptions, versionInt) {
		c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #PRA01 in"})
		return
	}

	entityFromStr := c.Params.ByName("entity")
	if entityFromStr == "" || !contains.ContainsStr(entityOptions, strings.ToLower(entityFromStr)) {
		c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #PRA02 in"})
		return
	}

	actionFromStr := c.Params.ByName("action")
	if actionFromStr == "" || !contains.ContainsStr(actionOptions, strings.ToLower(actionFromStr)) {
		c.JSON(http.StatusOK, gin.H{"gw_err": "Wrong argument #PRA03 in"})
		return
	}

}

func GwGetList(c *gin.Context) {

}

func GwGetIngridient(c *gin.Context) {

}
