package router

import (
    "app/internal/pkg/acme"
    "app/internal/server/http/mixed"
    "app/pkg/hook"
    "fmt"
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/spf13/viper"
    "golang.org/x/time/rate"
)

func Route(e *echo.Echo) {

    e.Use(middleware.Secure())
    e.Use(middleware.CORS())
    e.Use(middleware.Logger())

    // 调试中间件
    debug := viper.GetBool("app.debug")
    if !debug {
        e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
            DisableStackAll:   true,
            DisablePrintStack: true,
        }))
    }

    // 签名校验中间件
    e.Use(hook.SignatureWithConfig(hook.SignatureConfig{
        Salt: viper.GetString("app.salt"),
        Skipper: func(c echo.Context) bool {
            return true
        },
        PassRoute: []string{ // 无需签名接口
            "GET:/",
            "GET:/favicon.ico",
        },
    }))

    // JWT中间件
    jwtMiddle := hook.JWTWithConfig(hook.JWTConfig{
        SigningKey: []byte(viper.GetString("jwt.secret")),
        Claims:     &acme.UserInfoApi{},
        NeutralRoute: []string{ // 中立路由（给定则解析，不给定则跳过）
            "GET:/",
        },
    })

    // 限速中间件
    var rateMiddle = func(rate rate.Limit) echo.MiddlewareFunc {
        return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
            Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
                Rate: rate,
            }),
            IdentifierExtractor: func(context echo.Context) (string, error) {
                token := context.Request().Header.Get("Authorization") // 用户标识
                return token, nil
            },
            DenyHandler: func(context echo.Context, identifier string, err error) error {
                return &echo.HTTPError{
                    Code:     http.StatusTooManyRequests,
                    Message:  "请求过于频繁",
                    Internal: err,
                }
            },
        })
    }

    // WebSocket Route
    RWsBehavior.WsBehaviorRouter(e)
    RWsChat.WsChatRoute(e)
    RWsFive.WsFiveRoute(e)

    // 项目路由
    e.GET("/", mixed.CHome.GetHomepage, jwtMiddle, rateMiddle(1))             // 首页
    e.GET("/demo", mixed.CDemo.GetDemo, jwtMiddle, rateMiddle(10))            // Demo
    e.GET("/configure", mixed.CConfig.GetConfigure, jwtMiddle, rateMiddle(1)) // 获取配置

    e.Static("/node_modules", "public/node_modules")
    e.Static("/js", "public/js")
    e.Static("/css", "public/css")
    e.GET("/v/:html", func(c echo.Context) error {
        return c.File(fmt.Sprintf("public/view/%s.html", c.Param("html")))
    })
}
