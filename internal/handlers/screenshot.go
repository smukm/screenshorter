package handlers

import "github.com/gin-gonic/gin"

func (h Handler) Make(ctx *gin.Context) {

	filename := h.service.Screenshot.Make()

	ctx.JSON(200, filename)
}
