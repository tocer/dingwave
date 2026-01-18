package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, data)
}

func Error(c echo.Context, code int, message string) error {
	return c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}
