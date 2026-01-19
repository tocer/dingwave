package server

import (
	"dingtalk/internal/api"
	"dingtalk/internal/logger"
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

func New(db *gorm.DB, distFS embed.FS) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			logger.Info("%s %s %d %s", c.Request().Method, c.Request().RequestURI, c.Response().Status, time.Since(start))
			return err
		}
	})
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	handler := api.NewHandler(db)

	apiGroup := e.Group("/api")
	apiGroup.GET("/conversations/home", handler.GetConversationsHome)
	apiGroup.GET("/conversations", handler.GetConversations)
	apiGroup.GET("/conversations/:cid/messages", handler.GetConversationMessages)
	apiGroup.GET("/conversations/:cid/messages/search", handler.SearchConversationMessages)
	apiGroup.GET("/messages/search", handler.SearchMessagesGlobal)
	apiGroup.GET("/users", handler.GetUsers)
	apiGroup.GET("/users/search", handler.SearchUsers)

	distSubFS, _ := fs.Sub(distFS, "dist")
	e.GET("/*", echo.WrapHandler(http.FileServer(http.FS(distSubFS))))

	return e
}
