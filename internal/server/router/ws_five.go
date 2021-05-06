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
    Nickname   string          `json:"nickname"`    // 玩家昵称
    PlayerId   uint64          `json:"player_id"`   // 玩家ID，全场唯一
    RoomIndex  int             `json:"room_index"`  // 房间号
    Session    *melody.Session `json:"-"`           // 用户会话标识
    ChessColor string          `json:"chess_color"` // 棋色
    ChessStep  int             `json:"chess_step"`  // N手
}

// WebSocket five in a row
func (w *WsFive) WsFiveRoute(e *echo.Echo) {
    wsm := melody.New()

    // 所有玩家，通过会话找用户ID
    var sessionPlayers = make(map[*melody.Session]Player, 0)

    // 所有玩家，通过房间号找用户集合
    var roomPlayers []map[uint64]*melody.Session

    lock := new(sync.Mutex)
    e.GET("/ws/five", func(c echo.Context) error {
        _ = wsm.HandleRequest(c.Response().Writer, c.Request())
        return nil
    })

    colorMap := map[string]string{
        "black": "黑色方",
        "white": "白色方",
    }

    // 找房间坐下
    // 现有房间没有空位子则创建新房间
    var sitDown func(player *Player, times int64) (th int64)
    sitDown = func(player *Player, times int64) (th int64) {
        for roomIndex, items := range roomPlayers {
            if len(items) >= 2 {
                continue
            }
            th = int64(len(items) + 1)
            player.RoomIndex = roomIndex
            roomPlayers[roomIndex][player.PlayerId] = player.Session
            break
        }
        if player.RoomIndex < 0 {
            roomPlayers = append(roomPlayers, map[uint64]*melody.Session{})
            return sitDown(player, times+1)
        }
        return
    }

    // 返回指定房间的所有会话玩家
    var getSessionsByRoomIndex = func(roomIndex int) (players []*melody.Session) {
        for _, member := range roomPlayers[roomIndex] {
            players = append(players, member)
        }
        return
    }

    // 返回指定房间的所有玩家
    var getPlayerByRoomIndex = func(roomIndex int) (players []Player) {
        for _, member := range roomPlayers[roomIndex] {
            players = append(players, sessionPlayers[member])
        }
        return
    }

    // 处理连接
    wsm.HandleConnect(func(s *melody.Session) {
        lock.Lock()
        defer lock.Unlock()
        id, _ := variables.Snowflake.NextID() // 给当前会话生成一个ID
        sessionPlayers[s] = Player{
            Nickname:  tool.ToStr(id),
            PlayerId:  id,
            RoomIndex: -1,
            Session:   s,
        }
    })

    // 处理断开
    wsm.HandleDisconnect(func(s *melody.Session) {
        lock.Lock()
        defer lock.Unlock()
        me := sessionPlayers[s]
        // 对该房间的人进行广播.玩家下线
        msg := behavior.CBehavior.Message("offline", "", map[string]interface{}{
            "message": fmt.Sprintf("[%s] %s 下线了", colorMap[me.ChessColor], me.Nickname),
            "online":  len(roomPlayers[me.RoomIndex]) - 1,
        })
        _ = wsm.BroadcastMultiple([]byte(msg), getSessionsByRoomIndex(me.RoomIndex))
        delete(roomPlayers[me.RoomIndex], me.PlayerId)
        delete(sessionPlayers, s)
    })

    // 处理消息
    wsm.HandleMessage(func(s *melody.Session, msg []byte) {
        lock.Lock()
        defer lock.Unlock()
        args := &behavior.WebSocketRequest{}
        tool.JsonToInterface(string(msg), &args)

        if args.Behavior == "register" {
            me := sessionPlayers[s]
            me.Nickname = args.Arguments["nickname"].(string)
            args.Arguments["user_id"] = int64(me.PlayerId)

            // 找有空位的房间坐下
            th := sitDown(&me, 1)
            if th == 1 {
                me.ChessColor = "black"
            } else {
                me.ChessColor = "white"
            }
            sessionPlayers[s] = me

            // 返回注册信息
            args.Arguments["token"] = w.TokenCreator(args.Arguments)
            args.Arguments["room_number"] = me.RoomIndex + 1001
            behavior.CBehavior.WsResponse(s, args.Behavior, args.BehaviorId, args.Arguments)

            // 对该房间的人进行广播.玩家上线
            msg := behavior.CBehavior.Message("online", args.BehaviorId, map[string]interface{}{
                "message": fmt.Sprintf("[%s] %s 上线了", colorMap[me.ChessColor], me.Nickname),
                "online":  len(roomPlayers[me.RoomIndex]),
                "players": getPlayerByRoomIndex(me.RoomIndex),
            })
            _ = wsm.BroadcastMultiple([]byte(msg), getSessionsByRoomIndex(me.RoomIndex))
            return
        }

        _, err := w.ParseToken(args.Authorization)
        if err != nil {
            behavior.CBehavior.WsFailed(s, err.Error(), args.BehaviorId)
            return
        }
        return
    })
}
