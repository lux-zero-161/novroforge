import {GET} from '../modules/fetch.ts';
import {formatWingsBytes} from './wings-format.js';

const refreshInterval = 5000;
const maxConsoleLines = 500;

export function initWingsConsole() {
  const container = document.getElementById('wings-console');
  const output = document.getElementById('wings-console-output');
  const status = document.getElementById('wings-console-status');
  const serverState = document.getElementById('wings-server-state');
  const serverSuspended = document.getElementById('wings-server-suspended');
  const serverCpu = document.getElementById('wings-server-cpu');
  const serverMemory = document.getElementById('wings-server-memory');
  const serverDisk = document.getElementById('wings-server-disk');
  const serverUptime = document.getElementById('wings-server-uptime');
  const serverNetworkRx = document.getElementById('wings-server-network-rx');
  const serverNetworkTx = document.getElementById('wings-server-network-tx');
  if (
    !container ||
    !output ||
    !status ||
    !serverState ||
    !serverSuspended ||
    !serverCpu ||
    !serverMemory ||
    !serverDisk ||
    !serverUptime ||
    !serverNetworkRx ||
    !serverNetworkTx
  ) return;

  const initialMemory = serverMemory.textContent.split('/').map((value) => Number(value.trim()));
  if (initialMemory.length === 2 && initialMemory.every(Number.isFinite)) {
    serverMemory.textContent = `${formatWingsBytes(initialMemory[0])} / ${formatWingsBytes(initialMemory[1])}`;
  }
  for (const element of [serverDisk, serverNetworkRx, serverNetworkTx]) {
    const bytes = Number(element.textContent);
    if (Number.isFinite(bytes)) element.textContent = formatWingsBytes(bytes);
  }

  let consoleConnected = false;
  let reconnectDelay = 1000;
  let reconnectTimer;
  let stopped = false;
  let socket;
  let connect;

  const refresh = async () => {
    if (!document.hidden) {
      try {
        const response = await GET(container.dataset.logsUrl, {
          headers: {Accept: 'application/json'},
          cache: 'no-store',
          credentials: 'same-origin',
        });
        if (!response.ok) throw new Error(`Unexpected HTTP status ${response.status}`);

        const payload = await response.json();
        const details = payload && payload.details;
        const utilization = details && details.utilization;
        if (
          !payload ||
          typeof payload !== 'object' ||
          !Array.isArray(payload.logs) ||
          !payload.logs.every((line) => typeof line === 'string') ||
          !details ||
          typeof details.state !== 'string' ||
          typeof details.is_suspended !== 'boolean' ||
          !utilization ||
          !Number.isFinite(utilization.cpu_absolute) ||
          !Number.isFinite(utilization.memory_bytes) ||
          !Number.isFinite(utilization.memory_limit_bytes) ||
          !Number.isFinite(utilization.disk_bytes) ||
          !Number.isFinite(utilization.uptime) ||
          !Number.isFinite(utilization.network?.rx_bytes) ||
          !Number.isFinite(utilization.network?.tx_bytes)
        ) {
          throw new Error('Wings server update has an invalid shape');
        }

        const wasAtBottom = output.scrollHeight - output.scrollTop - output.clientHeight < 24;
        if (!consoleConnected) {
          output.textContent = payload.logs.length ? payload.logs.join('\n') : container.dataset.empty;
        }
        if (wasAtBottom) output.scrollTop = output.scrollHeight;
        serverState.textContent = details.state;
        serverSuspended.textContent = details.is_suspended ? container.dataset.suspended : container.dataset.notSuspended;
        serverCpu.textContent = `${utilization.cpu_absolute}`;
        serverMemory.textContent = `${formatWingsBytes(utilization.memory_bytes)} / ${formatWingsBytes(utilization.memory_limit_bytes)}`;
        serverDisk.textContent = formatWingsBytes(utilization.disk_bytes);
        serverUptime.textContent = `${utilization.uptime}`;
        serverNetworkRx.textContent = formatWingsBytes(utilization.network.rx_bytes);
        serverNetworkTx.textContent = formatWingsBytes(utilization.network.tx_bytes);
        status.textContent = consoleConnected ? container.dataset.live : container.dataset.updated;
      } catch (error) {
        console.error('Could not refresh Wings server status and console:', error);
        status.textContent = container.dataset.error;
      }
    }

    window.setTimeout(refresh, refreshInterval);
  };

  const scheduleReconnect = () => {
    if (stopped || typeof window.WebSocket !== 'function') return;
    reconnectTimer = window.setTimeout(() => connect(), reconnectDelay);
    reconnectDelay = Math.min(reconnectDelay * 2, 30000);
  };

  connect = () => {
    if (stopped || typeof window.WebSocket !== 'function') return;
    try {
      const url = new URL(container.dataset.websocketUrl, window.location.href);
      url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
      socket = new window.WebSocket(url);
    } catch (error) {
      console.error('Could not connect to Wings console websocket:', error);
      status.textContent = container.dataset.reconnecting;
      scheduleReconnect();
      return;
    }

    const activeSocket = socket;
    let openedAt = 0;
    activeSocket.addEventListener('open', () => {
      consoleConnected = true;
      openedAt = Date.now();
      status.textContent = container.dataset.live;
    });
    activeSocket.addEventListener('message', (message) => {
      if (activeSocket !== socket) return;
      try {
        const event = JSON.parse(message.data);
        if (event.event !== 'console output' || !Array.isArray(event.args) || typeof event.args[0] !== 'string') return;

        const wasAtBottom = output.scrollHeight - output.scrollTop - output.clientHeight < 24;
        const current = output.textContent === container.dataset.empty ? '' : output.textContent;
        output.textContent = `${current}${current && !current.endsWith('\n') ? '\n' : ''}${event.args[0]}`
          .split('\n')
          .slice(-maxConsoleLines)
          .join('\n');
        if (wasAtBottom) output.scrollTop = output.scrollHeight;
      } catch (error) {
        console.error('Could not parse Wings console event:', error);
      }
    });
    activeSocket.addEventListener('close', () => {
      if (activeSocket !== socket) return;
      consoleConnected = false;
      if (openedAt && Date.now() - openedAt >= 10000) reconnectDelay = 1000;
      status.textContent = container.dataset.reconnecting;
      scheduleReconnect();
    });
    activeSocket.addEventListener('error', () => activeSocket.close());
  };

  window.addEventListener('pagehide', () => {
    stopped = true;
    window.clearTimeout(reconnectTimer);
    socket?.close();
  }, {once: true});

  window.setTimeout(refresh, refreshInterval);
  connect();
}
