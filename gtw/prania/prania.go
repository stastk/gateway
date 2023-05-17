package prania

import "github.com/gin-gonic/gin"

// route for sending data to us
var PraniaPath = "/prania/:entity/:action/:id"

var entityOptions = []string{"list", "ingridient"}
var actionOptions = []string{"show"}

func GwGetList(c *gin.Context) {

}

func GwGetIngridient(c *gin.Context) {

}
