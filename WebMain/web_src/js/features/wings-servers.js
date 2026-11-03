import {GET} from '../modules/fetch.ts';
import {formatWingsBytes} from './wings-format.js';

const refreshInterval = 10000;
const maxConcurrentRequests = 4;

async function updateCard({card, state, cpu, memory, disk, suspended}) {
  try {
    const response = await GET(card.dataset.statusUrl, {
      headers: {Accept: 'application/json'},
      cache: 'no-store',
      credentials: 'same-origin',
    });
    if (!response.ok) throw new Error(`Unexpected HTTP status ${response.status}`);

    const details = await response.json();
    const utilization = details && details.utilization;
    if (
      !details ||
      typeof details.state !== 'string' ||
      typeof details.is_suspended !== 'boolean' ||
      !utilization ||
      !Number.isFinite(utilization.cpu_absolute) ||
      !Number.isFinite(utilization.memory_bytes) ||
      !Number.isFinite(utilization.memory_limit_bytes) ||
      !Number.isFinite(utilization.disk_bytes)
    ) {
      throw new Error('Wings server status has an invalid shape');
    }

    state.textContent = details.state;
    cpu.textContent = `${utilization.cpu_absolute}`;
    memory.textContent = `${formatWingsBytes(utilization.memory_bytes)} / ${formatWingsBytes(utilization.memory_limit_bytes)}`;
    disk.textContent = formatWingsBytes(utilization.disk_bytes);
    suspended.textContent = details.is_suspended ? card.dataset.suspended : card.dataset.notSuspended;
    card.removeAttribute('data-status-error');
  } catch (error) {
    console.error('Could not refresh Wings server overview status:', error);
    state.textContent = card.dataset.error;
    cpu.textContent = '-';
    memory.textContent = '-';
    disk.textContent = '-';
    suspended.textContent = '-';
    card.setAttribute('data-status-error', 'true');
  }
}

export function initWingsServers() {
  const cards = Array.from(document.querySelectorAll('[data-wings-server-card]'), (card) => ({
    card,
    searchText: `${card.querySelector('.header')?.textContent || ''} ${card.querySelector('code')?.textContent || ''}`.toLowerCase(),
    state: card.querySelector('[data-server-state]'),
    cpu: card.querySelector('[data-server-cpu]'),
    memory: card.querySelector('[data-server-memory]'),
    disk: card.querySelector('[data-server-disk]'),
    suspended: card.querySelector('[data-server-suspended]'),
  }));
  if (!cards.length) return;

  const filterInput = document.querySelector('[data-wings-server-filter]');
  const filterEmpty = document.querySelector('[data-wings-server-filter-empty]');
  if (filterInput && filterEmpty) {
    filterInput.addEventListener('input', () => {
      const query = filterInput.value.trim().toLowerCase();
      let matches = 0;
      for (const {card, searchText} of cards) {
        const isMatch = searchText.includes(query);
        card.hidden = !isMatch;
        if (isMatch) matches++;
      }
      filterEmpty.hidden = matches > 0;
    });
  }

  let stopped = false;
  let timer;

  const refresh = async () => {
    if (!document.hidden) {
      let nextCard = 0;
      const workers = Array.from({length: Math.min(maxConcurrentRequests, cards.length)}, async () => {
        while (nextCard < cards.length) {
          const currentCard = cards[nextCard++];
          await updateCard(currentCard);
        }
      });
      await Promise.all(workers);
    }
    if (!stopped) timer = window.setTimeout(refresh, refreshInterval);
  };

  window.addEventListener('pagehide', () => {
    stopped = true;
    window.clearTimeout(timer);
  }, {once: true});

  window.setTimeout(refresh, 0);
}
