document.getElementById("log").innerHTML = "";

const webSocket = new WebSocket("ws://127.0.0.1:8080/websocket/connect");

webSocket.onopen = e => {
  console.log(e, "Connection Initialized");
};

webSocket.onmessage = e => {
  console.log(e.data, "data received");
  document.getElementById("log").innerHTML += e.data;
  document.getElementById("log").innerHTML += "<br />";
};

webSocket.onclose = e => {
  console.log("Closing websocket connection");
};
