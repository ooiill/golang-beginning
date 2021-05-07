package mixed

import (
    "app/internal/pkg/acme"
    "app/pkg/handler"
    "github.com/labstack/echo/v4"
)

var CDemo Demo

type Demo struct {
    handler.Response
    acme.UserInfo
    acme.Acme
}

// Demo
func (cc *Demo) GetDemo(c echo.Context) error {
    _ = cc.ParseUserInfo(c)
    return cc.RS(c).ShowMessage("DONE")
}
