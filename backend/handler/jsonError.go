package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func JSONError(code int, ctx *gin.Context, e string) {
	ctx.JSON(code, gin.H{"error": fmt.Errorf(e)})
}
