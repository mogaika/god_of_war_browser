'use strict';

var wsStatusSocket;

function gowWsStatusConnect() {
    console.info("Trying to connect to ws server");
    if (window["WebSocket"]) {
        wsStatusSocket = new WebSocket("ws://" + document.location.host + "/ws/status");
        wsStatusSocket.onclose = function(evt) {
            console.info("SOCKET CLOSED", evt);
            wsStatusSocket = undefined;
            setTimeout(gowWsStatusConnect, 2 * 1000);
        };
        wsStatusSocket.onerror = function(evt) {
            console.error("SOCKET ERROR", evt);
            if (wsStatusSocket) {
                wsStatusSocket.close();
            }
            wsStatusSocket = undefined;
            setTimeout(gowWsStatusConnect, 2 * 1000);
        }
        wsStatusSocket.onmessage = function(evt) {
            var s = JSON.parse(evt.data);
            var $sp = $("#status-progress");
            var $st = $("#status-text");
            $st.text(s.Message);
            $sp.removeClass("info error progress");
            switch (s.Type) {
                case 0:
                    $sp.addClass("info");
                    $sp.width("100%");
                    break;
                case 1:
                    $sp.addClass("error");
                    $sp.width("100%");
                    break;
                case 2:
                    $sp.addClass("progress");
                    var progress = s.Progress;
                    if (progress > 1) {
                        progress = 1;
                    }
                    if (progress < 0) {
                        progress = 0;
                    }
                    $sp.width(progress * 100 + "%");
                    break;
            }
        };
    } else {
        console.warn("Your browser do not support websocket");
    }
}

$(document).ready(function() {
    $(window).on('beforeunload', function() {
        if (wsStatusSocket) {
            wsStatusSocket.close();
        }
    });
    gowWsStatusConnect();
});