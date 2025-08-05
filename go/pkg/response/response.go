package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(ctx *gin.Context, data interface{}) {
	resp := Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
	ctx.JSON(http.StatusOK, resp)
}

func Fail(ctx *gin.Context, code int, err error) {
	resp := Response{
		Code:    code,
		Message: err.Error(),
	}
	ctx.JSON(http.StatusInternalServerError, resp)
}

func New(ctx *gin.Context, data interface{}, err error) {
	if err != nil {
		Fail(ctx, -1, err)
	} else {
		Success(ctx, data)
	}
}
