'use strict';

bsw.configure({
    data: {
        ws: null,
        nickname: null,
        token: null,
        userId: null,
        online: 0,
        roomNumber: 0,
        blackPlayer: {},
        whitePlayer: {},
        willDown: 0
    },
    method: {
        inputNickname: function inputNickname() {
            var nickname = prompt("请输入你的昵称");
            if (nickname == null || nickname.length === 0) {
                return this.inputNickname();
            }
            return nickname;
        },
        drop: function drop(element) {
            element.mousedown(function (e) {
                var positionDiv = $(this).offset();
                var distenceX = e.pageX - positionDiv.left;
                var distenceY = e.pageY - positionDiv.top;
                $(document).mousemove(function (e) {
                    var x = e.pageX - distenceX;
                    var y = e.pageY - distenceY;
                    if (x < 0) {
                        x = 0;
                    } else if (x > $(document).width() - $('div').outerWidth(true)) {
                        x = $(document).width() - $('div').outerWidth(true);
                    }
                    if (y < 0) {
                        y = 0;
                    } else if (y > $(document).height() - $('div').outerHeight(true)) {
                        y = $(document).height() - $('div').outerHeight(true);
                    }
                    element.css({ 'left': x + 'px', 'top': y + 'px' });
                });
                $(document).mouseup(function () {
                    $(document).off('mousemove');
                });
            });
        }
    },
    logic: {
        ws: function ws(v) {
            v.nickname = v.inputNickname();
            v.ws = new ReconnectingWebSocket('ws://' + window.location.host + '/ws/five');
            v.ws.onopen = function () {
                console.log("成功连接服务器。");
                v.ws.send(JSON.stringify({
                    Behavior: "register",
                    Arguments: {
                        nickname: v.nickname
                    }
                }));
            };
            v.ws.onclose = function () {
                console.log("收到关闭请求，成功断开连接。");
            };

            v.ws.onmessage = function (msg) {
                var response = eval('(' + msg.data + ')');
                console.log("收到消息", response);
                if (response.Behavior === "register") {
                    var _ref = [response.Arguments.token, response.Arguments.user_id];
                    v.token = _ref[0];
                    v.userId = _ref[1];
                } else if (response.Behavior === "online") {
                    bsw.success(response.Arguments.message);
                    var _iteratorNormalCompletion = true;
                    var _didIteratorError = false;
                    var _iteratorError = undefined;

                    try {
                        for (var _iterator = response.Arguments.players[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
                            var member = _step.value;

                            if (member.chess_color === "black") {
                                v.blackPlayer = member;
                            } else {
                                v.whitePlayer = member;
                            }
                        }
                    } catch (err) {
                        _didIteratorError = true;
                        _iteratorError = err;
                    } finally {
                        try {
                            if (!_iteratorNormalCompletion && _iterator.return) {
                                _iterator.return();
                            }
                        } finally {
                            if (_didIteratorError) {
                                throw _iteratorError;
                            }
                        }
                    }
                } else if (response.Behavior === "offline") {
                    bsw.warning(response.Arguments.message);
                }
                if (typeof response.Arguments.room_number !== 'undefined') {
                    v.roomNumber = response.Arguments.room_number;
                }
                if (typeof response.Arguments.online !== 'undefined') {
                    v.online = response.Arguments.online;
                }
            };

            $(".chess").hover(function () {
                var isBlack = $(this).hasClass("black");
                $(".container").append('<div class="chess abs ' + (isBlack ? 'black' : 'white') + '"></div>');
                var offset = bsw.offset($(this));
                $("div.abs").css({ left: offset.left, top: offset.top });
            }, function () {
                $(".container div.abs").remove();
            });
        }
    }
});
