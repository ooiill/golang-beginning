/*!  */
/*! beginning - 1.0.0 - 2021-05-07 */
"use strict";bsw.configure({data:{ws:null,nickname:null,token:null,userId:null,online:0,roomNumber:0,blackPlayer:{},whitePlayer:{},myPlayer:{},willDown:0,nextColor:"black",chessHistory:{},isWin:!1},method:{inputNickname:function(){var a=prompt("请输入你的昵称");return null==a||0===a.length?this.inputNickname():a},cloneChess:function(){var a=this,b=bsw.offset($("div.data > div > div."+a.nextColor));$(".container > .abs").remove(),$(".container").append('<div class="chess abs '+a.nextColor+'"></div>');var c=$(".container > div.abs");c.css({left:b.left,top:b.top}),a.willDown=0,a.drop(c)},drop:function(a){var b=this;a.hasClass(b.myPlayer.chess_color)&&a.mousedown(function(c){$(document).mousemove(function(d){var e=d.pageX-c.offsetX,f=d.pageY-c.offsetY;a.css({left:e,top:f});var g=Math.floor((d.pageX-70)/50),h=Math.floor((d.pageY-70)/50);if(d.pageX>45&&d.pageX<70?g=0:(d.pageX-70)%50>25&&(g+=1),d.pageY>45&&d.pageY<70?h=0:(d.pageY-70)%50>25&&(h+=1),g<0||g>14||h<0||h>14)b.willDown=0;else{var i=15*h+(g+1);b.chessHistory[i]&&(i=0),b.willDown=i}b.isWin||b.ws.send(JSON.stringify({Authorization:b.token,Behavior:"move",Arguments:{left:e,top:f,willDown:b.willDown}}))}),$(document).mouseup(function(){if($(document).off("mousemove"),$(document).off("mouseup"),b.willDown>0){if(b.isWin)return;b.ws.send(JSON.stringify({Authorization:b.token,Behavior:"chessDown",Arguments:{willDown:b.willDown}}))}})})},resetPlayers:function(a){var b=this,c=!0,d=!1,e=void 0;try{for(var f,g=a[Symbol.iterator]();!(c=(f=g.next()).done);c=!0){var h=f.value;"black"===h.chess_color?b.blackPlayer=h:b.whitePlayer=h,h.player_id===b.userId&&(b.myPlayer=h)}}catch(a){d=!0,e=a}finally{try{!c&&g.return&&g.return()}finally{if(d)throw e}}}},logic:{ws:function ws(v){v.nickname=v.inputNickname(),v.ws=new ReconnectingWebSocket("ws://"+window.location.host+"/ws/five"),v.ws.onopen=function(){console.log("成功连接服务器。"),v.ws.send(JSON.stringify({Behavior:"register",Arguments:{nickname:v.nickname}}))},v.ws.onclose=function(){console.log("收到关闭请求，成功断开连接。")},v.ws.onmessage=function(msg){var response=eval("("+msg.data+")"),args=response.Arguments;if("move"!==response.Behavior&&console.log("收到消息",response),"register"===response.Behavior){var _ref=[args.token,args.user_id];v.token=_ref[0],v.userId=_ref[1]}else if("online"===response.Behavior)bsw.success(args.message),v.nextColor=args.nextColor,v.resetPlayers(args.players),2===args.players.length&&v.cloneChess(),v.chessHistory=args.history;else if("offline"===response.Behavior)bsw.warning(args.message);else if("move"===response.Behavior){var abs=$(".container > div.abs");abs.css({left:args.left,top:args.top}),v.willDown=args.willDown}else"chessDown"===response.Behavior&&(v.chessHistory=args.history,v.nextColor=args.nextColor,v.resetPlayers(args.players),v.cloneChess(),args.win.length>0&&(v.isWin=!0,args.win===v.myPlayer.chess_color?bsw.confirm("success","你赢了，继续保持哟~",0):bsw.confirm("warning","你输了，要努力了哟~",0)));"undefined"!=typeof args.room_number&&(v.roomNumber=args.room_number),"undefined"!=typeof args.online&&(v.online=args.online)}}}});