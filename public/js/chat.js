$(function () {
    // 输入用户昵称
    let inputNickname = function () {
        let nickname = prompt("请输入你的昵称");
        if (nickname == null || nickname.length === 0) {
            return inputNickname();
        }
        return nickname;
    }
    let nickname = inputNickname()
    $("sup#name").html(nickname);

    let ws = new ReconnectingWebSocket(`ws://${window.location.host}/ws/chat`);
    ws.onopen = function () {
        console.log("成功连接服务器。");
        ws.send(JSON.stringify({
            Behavior: "register",
            Arguments: {
                nickname: nickname
            }
        }))
    };
    ws.onclose = function () {
        console.log("收到关闭请求，成功断开连接。");
    };

    let token, user_id;
    let chatDiv = $("div#chat")
    ws.onmessage = function (msg) {
        let response = eval(`(${msg.data})`)
        console.log("收到消息", response)
        if (response.Behavior === "register") {
            [token, user_id] = [response.Arguments.token, response.Arguments.user_id];
        } else if (response.Behavior === "online") {
            chatDiv.append(`<p class="online"><b> ${response.Arguments.nickname} 上线了</b></p>`)
            $("span#counter").html(response.Arguments.online);
        } else if (response.Behavior === "offline") {
            chatDiv.append(`<p class="offline"><b> ${response.Arguments.nickname} 下线了</b></p>`)
            $("span#counter").html(response.Arguments.online);
        } else if (response.Behavior === "chat") {
            chatDiv.append(`<p><b>${response.Arguments.time}</b> ${response.Arguments.from}: ${response.Arguments.message}</p>`)
        }
        chatDiv.animate({scrollTop: chatDiv.prop("scrollHeight")}, 200);
    };

    let text = $("input#text")
    text.on("keydown", function (e) {
        if (e.keyCode === 13 && text.val() !== "") {
            ws.send(JSON.stringify({
                Authorization: token,
                Behavior: "chat",
                Arguments: {
                    message: text.val()
                },
                Time: Math.floor((new Date()).getTime() / 1000)
            }))
            text.val("");
        }
    })
});