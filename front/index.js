import { Terminal } from '@xterm/xterm';
import { AttachAddon } from '@xterm/addon-attach';
import { FitAddon } from '@xterm/addon-fit';

import '@xterm/xterm/css/xterm.css';

const terminal = new Terminal();
const fitAddon = new FitAddon();
terminal.loadAddon(fitAddon);
terminal.open(document.getElementById('terminal'));
fitAddon.fit();

const webSocketProtocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
const webSocket = new WebSocket(
  `${webSocketProtocol + window.location.host}/web-socket/ssh`,
);

const sendSize = () => {
  const windowSize = { high: terminal.rows, width: terminal.cols };
  const blob = new Blob([JSON.stringify(windowSize)], {
    type: 'application/json',
  });
  webSocket.send(blob);
};

const resizeScreen = () => {
  fitAddon.fit();
  sendSize();
};

webSocket.onopen = sendSize;

webSocket.onmessage = (event) => {
  try {
    const message = JSON.parse(event.data);
    if (message.type === 'info') {
      terminal.write(`\r\n${message.data}`);
    } else if (message.type === 'error') {
      terminal.write(`\r\nError: ${message.data}`);
    } else {
      terminal.write(`\r\nUnknown message type: ${message.type}`);
    }
  } catch (error) {
    // pass
  }
};

webSocket.addEventListener('error', (event) => {
  terminal.write(`\r\nWebSocket error: ${event.message}`);
});

webSocket.addEventListener('close', () => {
  terminal.write('\r\nConnection closed, please refresh the page to reconnect.');
});

window.addEventListener('resize', resizeScreen, false);

const attachAddon = new AttachAddon(webSocket);
terminal.loadAddon(attachAddon);
