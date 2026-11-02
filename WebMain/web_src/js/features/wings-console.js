import {GET} from '../modules/fetch.ts';

const refreshInterval = 5000;

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
        output.textContent = payload.logs.length ? payload.logs.join('\n') : container.dataset.empty;
        if (wasAtBottom) output.scrollTop = output.scrollHeight;
        serverState.textContent = details.state;
        serverSuspended.textContent = details.is_suspended ? container.dataset.suspended : container.dataset.notSuspended;
        serverCpu.textContent = `${utilization.cpu_absolute}`;
        serverMemory.textContent = `${utilization.memory_bytes} / ${utilization.memory_limit_bytes}`;
        serverDisk.textContent = `${utilization.disk_bytes}`;
        serverUptime.textContent = `${utilization.uptime}`;
        serverNetworkRx.textContent = `${utilization.network.rx_bytes}`;
        serverNetworkTx.textContent = `${utilization.network.tx_bytes}`;
        status.textContent = container.dataset.updated;
      } catch (error) {
        console.error('Could not refresh Wings server status and console:', error);
        status.textContent = container.dataset.error;
      }
    }

    window.setTimeout(refresh, refreshInterval);
  };

  window.setTimeout(refresh, refreshInterval);
}
