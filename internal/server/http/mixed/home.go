package mixed

import (
    "app/internal/pkg/acme"
    "app/pkg/handler"
    "github.com/labstack/echo/v4"
    "github.com/spf13/viper"
    "net/http"
)

var CHome Home

type Home struct {
    handler.Response
    acme.UserInfo
    acme.Acme
}

// Homepage
func (cc *Home) GetHomepage(c echo.Context) error {
    usr := cc.ParseUserInfo(c)
    return c.JSON(http.StatusOK, cc.RS().ShowOkay(echo.Map{
        "version": viper.GetString("app.version"),
        "speech":  viper.GetString("app.name"),
        "uid":     usr.UID,
    }))
}