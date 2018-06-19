'use strict';

function gowWsStatusConnect() {
	console.info("Trying to connect to ws server");
	if (window["WebSocket"]) {
        var conn = new WebSocket("ws://" + document.location.host + "/ws/status");
        conn.onclose = function (evt) {
            console.info("Connection closed");
			setTimeout(gowWsStatusConnect, 3*1000);
        };
        conn.onmessage = function (evt) {
			console.log("msg", evt.data);
        };
    } else {
        console.warn("Your browser do not support websocket");
    }
}

$(document).ready(function() {
	gowWsStatusConnect();
});
