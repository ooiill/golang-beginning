package router

import (
    "app/internal/pkg/acme"
    "app/internal/server/http/behavior"
    "app/internal/server/variables"
    tool "app/pkg/acme"
    "fmt"
    "github.com/golang-module/carbon"
    "github.com/labstack/echo/v4"
    "gopkg.in/olahol/melody.v1"
    "sync"
)

var RWsLandlord WsLandlord

type WsLandlord struct {
    acme.UserInfo
}

type Player struct {
    Nickname   string // 玩家昵称
    PlayerId   uint64 // 玩家ID，全场唯一
    RoomIndex int64  // 房间号
}

// WebSocket landlord
func (w *WsLandlord) WsLandlordRoute(e *echo.Echo) {
    wsm := melody.New()

    // 所有玩家，通过会话索引
    var sessionPlayers = make(map[*melody.Session]Player, 0)
    // 所有玩家，通过ID索引
    var idPlayers = make(map[uint64]*melody.Session, 0)
    // 所有玩家.房间索引
    var roomPlayers [5][]Player // 初始化只创建 5 个房间

    // 当前各房间对应剩余坐席数
    var roomSeats = []int64{
        3, // 房间 1
        3, // 房间 2
        3, // 房间 3
        3, // 房间 4
        3, // 房间 5
    }

    lock := new(sync.Mutex)
    e.GET("/ws/landlord", func(c echo.Context) error {
        _ = wsm.HandleRequest(c.Response().Writer, c.Request())
        return nil
    })

    // 查找空房间
    var findSeats = func() int64 {
        // 找有空位的房间坐下
        var theSeats int64
        for roomIndex, seats := range roomSeats {
            if seats == 0 {
                continue
            }
            theSeats = int64(roomIndex + 1)
            roomSeats[roomIndex] -= 1
            break
        }
        return theSeats
    }

    // 返回指定房间的所有会话玩家
    var getSessionPlayerByRoomIndex = func(roomPlayers []Player) (players []*melody.Session) {
        for _, member := range roomPlayers {
            players = append(players, idPlayers[member.PlayerId])
        }
        return
    }

    // 处理连接
    wsm.HandleConnect(func(s *melody.Session) {
        lock.Lock()
        defer lock.Unlock()
        id, _ := variables.Snowflake.NextID()
        sessionPlayers[s] = Player{
            Nickname: tool.ToStr(id),
            PlayerId: id,
        }
    })

    // 处理断开
    wsm.HandleDisconnect(func(s *melody.Session) {
        lock.Lock()
        defer lock.Unlock()
        // 对该房间的人进行广播.玩家上线
        msg := behavior.CBehavior.Message("offline", "", echo.Map{
            "message": fmt.Sprintf("%s 下线了", sessionPlayers[s].Nickname),
            "people":  len(roomPlayers[sessionPlayers[s].RoomIndex]),
        })
        _ = wsm.BroadcastMultiple([]byte(msg), getSessionPlayerByRoomIndex(roomPlayers[sessionPlayers[s].RoomIndex]))
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
            args.Arguments["user_id"] = int64(tool.ToInt(me.PlayerId))
            if _, ok := args.Arguments["nickname"]; !ok {
                args.Arguments["nickname"] = me.Nickname
            }

            // 找有空位的房间坐下
            me.RoomIndex = findSeats()
            me.Nickname = args.Arguments["nickname"].(string)

            // 没找到房间则提示
            if me.RoomIndex == 0 {
                behavior.CBehavior.WsFailed(s, "暂时没有空位置，等会再来玩吧~", "")
                return
            }

            // 进入房间
            roomPlayers[me.RoomIndex] = append(roomPlayers[me.RoomIndex], me)
            idPlayers[me.PlayerId] = s

            // 返回注册信息
            args.Arguments["token"] = w.TokenCreator(args.Arguments)
            args.Arguments["room"] = me.RoomIndex
            behavior.CBehavior.WsResponse(s, args.Behavior, args.BehaviorId, args.Arguments)

            // 对该房间的人进行广播.玩家上线
            args.Arguments["message"] = fmt.Sprintf("%s 上线了", me.Nickname)
            args.Arguments["people"] = len(roomPlayers[me.RoomIndex])
            msg := behavior.CBehavior.Message("online", args.BehaviorId, args.Arguments)
            _ = wsm.BroadcastMultiple([]byte(msg), getSessionPlayerByRoomIndex(roomPlayers[me.RoomIndex]))
            return
        }

        usr, err := w.ParseToken(args.Authorization)
        if err != nil {
            behavior.CBehavior.WsFailed(s, err.Error(), args.BehaviorId)
            return
        }

        go func() { // 单用户指令
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
