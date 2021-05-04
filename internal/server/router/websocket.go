package router

import (
    "app/internal/entity"
    "app/internal/pkg/acme"
    "app/internal/server/http/behavior"
    "app/internal/server/variables"
    tool "app/pkg/acme"
    "app/pkg/handler"
    "fmt"
    "github.com/golang-module/carbon"
    "github.com/labstack/echo/v4"
    "gopkg.in/olahol/melody.v1"
    "net/http"
    "sync"
    "time"
)

var RWebSocket WebSocket
var Debug = false

type WebSocket struct {
    handler.Response
    acme.UserInfo
    acme.Acme
}

type Gophers struct {
    OnlineTime int64
    Nickname   string
}

// WebSocket
func (w *WebSocket) WebSocketRoute(e *echo.Echo) {
    wsm := melody.New()
    gophers := make(map[*melody.Session]Gophers, 0)
    lock := new(sync.Mutex)
    counter := 0
    wsm.Upgrader.CheckOrigin = func(r *http.Request) bool {
        return true
    }
    e.GET("/ws/behavior", func(c echo.Context) error {
        _ = wsm.HandleRequest(c.Response().Writer, c.Request())
        return nil
    })

    // 处理连接
    wsm.HandleConnect(func(s *melody.Session) {
        lock.Lock()
        defer lock.Unlock()
        id, _ := variables.Snowflake.NextID()
        gophers[s] = Gophers{
            OnlineTime: carbon.Now().ToTimestampWithSecond(),
            Nickname:   tool.ToStr(id),
        }
        counter += 1
        if Debug {
            tool.PrintVar(fmt.Sprintf("客户端连接，连接后的用户数为：%d", counter))
        }
        if counter == 1 { // 全服行为监听
            go func() {
                for {
                    allBehavior := variables.Rds.BRPop(time.Minute, entity.CK4AllRealTimeBehavior()).Val()
                    if len(allBehavior) >= 2 {
                        arguments := &acme.WebSocketResponse{}
                        tool.JsonToInterface(allBehavior[1], &arguments)
                        if Debug {
                            tool.PrintVar(fmt.Sprintf("收到全服队列消息：%+v", arguments))
                        }

                        // 不广播历史消息
                        t := carbon.Now().ToTimestampWithSecond()
                        if v, ok := gophers[s]; ok {
                            t = v.OnlineTime
                        }
                        if arguments.Time >= t {
                            msg := behavior.CBehavior.Message(arguments.Behavior, arguments.BehaviorId, arguments.Arguments)
                            _ = wsm.Broadcast([]byte(msg))
                        }
                    }
                    if counter <= 0 {
                        return
                    }
                }
            }()
        }
    })

    // 处理断开
    wsm.HandleDisconnect(func(s *melody.Session) {
        lock.Lock()
        defer lock.Unlock()
        msg := behavior.CBehavior.Message("offline", "", echo.Map{
            "nickname": gophers[s].Nickname,
            "counter":  counter - 1,
        })
        _ = wsm.Broadcast([]byte(msg))
        delete(gophers, s)
        counter -= 1
        if Debug {
            tool.PrintVar(fmt.Sprintf("客户端断开连接，断开后的用户数为：%d", counter))
        }
    })

    // 处理消息
    wsm.HandleMessage(func(s *melody.Session, msg []byte) {
        lock.Lock()
        defer lock.Unlock()
        args := &behavior.WebSocketRequest{}
        tool.JsonToInterface(string(msg), &args)

        if Debug {
            tool.PrintVar(fmt.Sprintf("收到消息：%+v", args))
        }
        if args.Behavior == "ping" {
            behavior.CBehavior.WsResponse(s, "pong", args.BehaviorId, nil)
            return
        }

        // TODO demo.register
        if args.Behavior == "register" {
            args.Arguments["user_id"] = gophers[s].Nickname
            if _, ok := args.Arguments["nickname"]; !ok {
                args.Arguments["nickname"] = args.Arguments["user_id"]
            }

            user := gophers[s]
            user.Nickname = tool.ToStr(args.Arguments["nickname"])
            gophers[s] = user

            // 返回注册信息
            args.Arguments["user_id"] = int64(tool.ToInt(args.Arguments["user_id"]))
            args.Arguments["token"] = w.TokenCreator(args.Arguments)
            behavior.CBehavior.WsResponse(s, args.Behavior, args.BehaviorId, args.Arguments)
            // 广播用户上线
            args.Arguments["counter"] = counter
            msg := behavior.CBehavior.Message("online", args.BehaviorId, args.Arguments)
            _ = wsm.Broadcast([]byte(msg))
            return
        }

        usr, err := w.ParseToken(args.Authorization)
        if err != nil {
            behavior.CBehavior.WsFailed(s, err.Error(), args.BehaviorId)
            return
        }

        if args.Behavior == "behavior:userRealTime" { // 单用户行为监听
            go func() {
                for {
                    userBehavior := variables.Rds.BRPop(time.Minute, entity.CK4UserRealTimeBehavior(usr.UID)).Val()
                    if len(userBehavior) >= 2 {
                        arguments := &acme.WebSocketResponse{}
                        tool.JsonToInterface(userBehavior[1], &arguments)
                        if Debug {
                            tool.PrintVar(fmt.Sprintf("收到单用户队列消息：%+v", arguments))
                        }
                        behavior.CBehavior.WsResponse(s, arguments.Behavior, arguments.BehaviorId, arguments.Arguments)
                    }
                    if t, ok := gophers[s]; ok && t.OnlineTime > 0 {
                        return
                    }
                }
            }()
        }

        go func() { // 单用户指令
            // TODO demo.chat
            if args.Behavior == "chat" {
                args.Arguments["from"] = usr.Nickname
                args.Arguments["from_id"] = usr.UID
                args.Arguments["time"] = carbon.CreateFromTimestamp(args.Time).ToFormatString("m/d H:i:s")
                msg := behavior.CBehavior.Message(args.Behavior, args.BehaviorId, args.Arguments)
                _ = wsm.Broadcast([]byte(msg))
            }
        }()
        return
    })
}
