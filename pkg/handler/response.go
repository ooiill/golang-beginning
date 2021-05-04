package handler

import (
    "github.com/golang-module/carbon"
    "github.com/labstack/echo/v4"
)

var ResponseHandler Response

type Response struct {
}

type ResponseStruct struct {
    responseHttpCode int
    responseCode     int
    responseMessage  string
    responseData     interface{}
}

func (r *Response) RS() *ResponseStruct {
    return &ResponseStruct{}
}

// error interface
func (rs *ResponseStruct) Error() string {
    return rs.responseMessage
}

// 设置响应错误码
func (rs *ResponseStruct) SetHttpCode(code int) *ResponseStruct {
    rs.responseHttpCode = code
    return rs
}

// 获取响应错误码
func (rs *ResponseStruct) GetHttpCode() int {
    return rs.responseHttpCode
}

// 设置逻辑错误码
func (rs *ResponseStruct) SetCode(code int) *ResponseStruct {
    rs.responseCode = code
    return rs
}

// 获取逻辑错误码
func (rs *ResponseStruct) GetCode() int {
    return rs.responseCode
}

// 设置响应消息
func (rs *ResponseStruct) SetMessage(message string) *ResponseStruct {
    rs.responseMessage = message
    return rs
}

// 获取响应消息
func (rs *ResponseStruct) GetMessage() string {
    return rs.responseMessage
}

// 设置响应数据
func (rs *ResponseStruct) SetData(data interface{}) *ResponseStruct {
    rs.responseData = data
    return rs
}

// 获取响应数据
func (rs *ResponseStruct) GetData() interface{} {
    if rs.responseData == nil {
        rs.responseData = echo.Map{}
    }
    return rs.responseData
}

// 追加响应数据
func (rs *ResponseStruct) AppendData(field string, value interface{}) *ResponseStruct {
    if rs.responseData == nil {
        rs.responseData = echo.Map{field: value}
    } else {
        res := rs.responseData.(echo.Map)
        res[field] = value
        rs.responseData = res
    }
    return rs
}

// 向响应结构中追加错误详情
func (rs *ResponseStruct) ErrorDetail(err interface{}) *ResponseStruct {
    switch msg := err.(type) {
    case string:
        _ = rs.AppendData("err_detail", msg)
    case error:
        _ = rs.AppendData("err_detail", msg.Error())
    case echo.Map:
    case map[string]interface{}:
        _ = rs.AppendData("err_detail", msg)
    }
    return rs
}

// 获取响应结构
func (rs *ResponseStruct) GetStruct() interface{} {
    if rs.responseData == nil {
        rs.responseData = echo.Map{}
    }
    result := echo.Map{
        "code":    rs.GetCode(),
        "message": rs.GetMessage(),
        "time":    carbon.Now().ToTimestampWithSecond(),
        "data":    rs.GetData(),
    }
    // tool.PrintVar(fmt.Sprintf("接口返回：%s", tool.InterfaceToJson(result)))
    return result
}

// 响应错误
func (rs *ResponseStruct) ShowError(message string, err interface{}) interface{} {
    return rs.SetMessage(message).ErrorDetail(err).GetStruct()
}

// 响应提示
func (rs *ResponseStruct) ShowMessage(message string) interface{} {
    return rs.SetMessage(message).GetStruct()
}

// 响应数据
func (rs *ResponseStruct) ShowOkay(data interface{}) interface{} {
    return rs.SetData(data).GetStruct()
}
