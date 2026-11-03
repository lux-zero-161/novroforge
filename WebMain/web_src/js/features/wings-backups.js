import {GET} from '../modules/fetch.ts';

const refreshInterval = 5000;
const backupStatuses = ['pending', 'completed', 'failed'];
const restoreStatuses = ['none', 'pending', 'completed', 'failed'];

export function initWingsBackups() {
  const container = document.getElementById('wings-backups');
  const status = document.getElementById('wings-backup-status');
  if (!container || !status) return;

  const rows = new Map(
    Array.from(container.querySelectorAll('[data-backup-id]'), (row) => [row.dataset.backupId, row]),
  );
  const filterInput = container.querySelector('[data-wings-backup-filter]');
  const filterEmpty = container.querySelector('[data-wings-backup-filter-empty]');
  const filterRows = () => {
    if (!filterInput || !filterEmpty) return;
    const query = filterInput.value.trim().toLowerCase();
    let matches = 0;
    for (const row of rows.values()) {
      const searchText = `${row.cells[0].textContent} ${row.querySelector('[data-backup-status]').textContent}`.toLowerCase();
      const isMatch = searchText.includes(query);
      row.hidden = !isMatch;
      if (isMatch) matches++;
    }
    filterEmpty.hidden = matches > 0;
  };
  filterInput?.addEventListener('input', filterRows);
  filterRows();

  const statusLabels = {
    pending: container.dataset.statusPending,
    completed: container.dataset.statusCompleted,
    failed: container.dataset.statusFailed,
  };
  const restoreStatusLabels = {
    pending: container.dataset.restoreStatusPending,
    completed: container.dataset.restoreStatusCompleted,
    failed: container.dataset.restoreStatusFailed,
  };

  const refresh = async () => {
    if (!document.hidden && rows.size) {
      try {
        const response = await GET(container.dataset.statusUrl, {
          headers: {Accept: 'application/json'},
          cache: 'no-store',
          credentials: 'same-origin',
        });
        if (!response.ok) throw new Error(`Unexpected HTTP status ${response.status}`);

        const payload = await response.json();
        if (
          !Array.isArray(payload) ||
          !payload.every((backup) =>
            backup &&
            typeof backup.uuid === 'string' &&
            backupStatuses.includes(backup.status) &&
            typeof backup.size === 'string' &&
            /^(0|[1-9]\d*)$/.test(backup.size) &&
            restoreStatuses.includes(backup.restore_status),
          )
        ) {
          throw new Error('Wings backup status response has an invalid shape');
        }

        for (const backup of payload) {
          const row = rows.get(backup.uuid);
          if (!row) continue;

          row.querySelector('[data-backup-status]').textContent = statusLabels[backup.status];
          row.querySelector('[data-backup-restore-status]').textContent = backup.restore_status === 'none' ?
            '' :
            ` | ${restoreStatusLabels[backup.restore_status]}`;
          row.querySelector('[data-backup-size]').textContent = backup.size;
          row.querySelector('[data-backup-download-link]').hidden = !(
            backup.status === 'completed' && container.dataset.backupDownloadEnabled === 'true'
          );
          row.querySelector('[data-backup-restore-form]').hidden = !(
            container.dataset.backupManagementEnabled === 'true' &&
            backup.status === 'completed' &&
            container.dataset.callbacksConfigured === 'true' &&
            backup.restore_status !== 'pending'
          );
          row.querySelector('[data-backup-delete-form]').hidden = (
            container.dataset.backupManagementEnabled !== 'true' ||
            backup.status === 'pending' || backup.restore_status === 'pending'
          );
        }
        filterRows();
        status.textContent = '';
      } catch (error) {
        console.error('Could not refresh Wings backup statuses:', error);
        status.textContent = container.dataset.error;
      }
    }

    window.setTimeout(refresh, refreshInterval);
  };

  window.setTimeout(refresh, refreshInterval);
}
