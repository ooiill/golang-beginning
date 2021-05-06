package router

import (
    "app/internal/pkg/acme"
    "app/internal/server/http/behavior"
    "app/internal/server/variables"
    tool "app/pkg/acme"
    "fmt"
    "github.com/labstack/echo/v4"
    "gopkg.in/olahol/melody.v1"
    "sync"
)

var RWsFive WsFive

type WsFive struct {
    acme.UserInfo
}

type Player struct {
    Nickname  string          // 玩家昵称
    PlayerId  uint64          // 玩家ID，全场唯一
    RoomIndex int             // 房间号
    Session   *melody.Session // 用户会话标识
}

// WebSocket five in a row
func (w *WsFive) WsFiveRoute(e *echo.Echo) {
    wsm := melody.New()

    // 所有玩家，通过会话找用户ID
    var sessionPlayers = make(map[*melody.Session]uint64, 0)

    // 所有玩家，通过用户ID找用户信息
    var idPlayers = make(map[uint64]Player, 0)

    // 所有玩家，通过房间号找用户集合
    var roomPlayers []map[uint64]Player

    lock := new(sync.Mutex)
    e.GET("/ws/five", func(c echo.Context) error {
        _ = wsm.HandleRequest(c.Response().Writer, c.Request())
        return nil
    })

    // 找房间坐下
    // 现有房间没有空位子则创建新房间
    var sitDown func(player *Player, times int64)
    sitDown = func(player *Player, times int64) {
        for roomIndex, items := range roomPlayers {
            if len(items) >= 2 {
                continue
            }
            player.RoomIndex = roomIndex
            roomPlayers[roomIndex][player.PlayerId] = *player
            break
        }
        if player.RoomIndex < 0 {
            roomPlayers = append(roomPlayers, map[uint64]Player{})
            sitDown(player, times+1)
        }
    }

    // 返回指定房间的所有会话玩家
    var getSessionsByRoomIndex = func(roomIndex int) (players []*melody.Session) {
        for _, member := range roomPlayers[roomIndex] {
            players = append(players, member.Session)
        }
        return
    }

    // 处理连接
    wsm.HandleConnect(func(s *melody.Session) {
        lock.Lock()
        defer lock.Unlock()
        sessionPlayers[s], _ = variables.Snowflake.NextID() // 给当前会话生成一个ID
    })

    // 处理断开
    wsm.HandleDisconnect(func(s *melody.Session) {
        lock.Lock()
        defer lock.Unlock()
        // 对该房间的人进行广播.玩家上线

        delete(sessionPlayers, s)
    })

    // 处理消息
    wsm.HandleMessage(func(s *melody.Session, msg []byte) {
        lock.Lock()
        defer lock.Unlock()
        args := &behavior.WebSocketRequest{}
        tool.JsonToInterface(string(msg), &args)

        if args.Behavior == "register" {
            me := Player{
                Nickname:  args.Arguments["nickname"].(string),
                PlayerId:  sessionPlayers[s],
                RoomIndex: -1,
                Session:   s,
            }
            idPlayers[me.PlayerId] = me
            args.Arguments["user_id"] = int64(me.PlayerId)

            // 找有空位的房间坐下
            sitDown(&me, 1)

            // 返回注册信息
            args.Arguments["token"] = w.TokenCreator(args.Arguments)
            args.Arguments["room"] = me.RoomIndex + 1001
            behavior.CBehavior.WsResponse(s, args.Behavior, args.BehaviorId, args.Arguments)

            // 对该房间的人进行广播.玩家上线
            msg := behavior.CBehavior.Message("online", args.BehaviorId, map[string]interface{}{
                "message": fmt.Sprintf("%s 上线了", me.Nickname),
                "people":  len(roomPlayers[me.RoomIndex]),
            })
            _ = wsm.BroadcastMultiple([]byte(msg), getSessionsByRoomIndex(me.RoomIndex))
            return
        }

        _, err := w.ParseToken(args.Authorization)
        if err != nil {
            behavior.CBehavior.WsFailed(s, err.Error(), args.BehaviorId)
            return
        }

        go func() { // 单用户指令

        }()
        return
    })
}
