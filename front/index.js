import { Terminal } from 'xterm';
import { AttachAddon } from 'xterm-addon-attach';

const terminal = new Terminal();
terminal.open(document.getElementById('terminal'));

const webSocket = new WebSocket("ws://127.0.0.1:8080/ws");

const attachAddon = new AttachAddon(webSocket);
terminal.loadAddon(attachAddon);
