import { Terminal } from 'xterm';
import { AttachAddon } from 'xterm-addon-attach';
import { FitAddon } from 'xterm-addon-fit';

const terminal = new Terminal();
const fitAddon = new FitAddon();
terminal.loadAddon(fitAddon);
terminal.open(document.getElementById('terminal'));
fitAddon.fit();

const webSocket = new WebSocket("ws://127.0.0.1:8080/web-socket/ssh");

const sendSize = () => {
  const windowSize = {high: terminal.rows, width: terminal.cols};
  const blob = new Blob([JSON.stringify(windowSize)], {type : 'application/json'});
  webSocket.send(blob);
}

webSocket.onopen = sendSize;

const resizeScreen = () => {
  fitAddon.fit();
  sendSize();
}
window.addEventListener('resize', resizeScreen, false);

// terminal.onResize(event => {
//   const windowSize = {high: event.rows, width: event.cols};
//   const blob = new Blob([JSON.stringify(windowSize)], {type : 'application/json'});
//   webSocket.send(blob);
// })

const attachAddon = new AttachAddon(webSocket);
terminal.loadAddon(attachAddon);
