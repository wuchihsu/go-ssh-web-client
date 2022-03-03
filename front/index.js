import { Terminal } from 'xterm';

const term = new Terminal();
term.open(document.getElementById('terminal'));
term.write('Heeeeeeello from \x1B[1;3;31mxterm.js\x1B[0m $ ');
