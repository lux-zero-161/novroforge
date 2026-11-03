export function initWingsAdminServers() {
  const filterInput = document.querySelector('[data-wings-admin-server-filter]');
  const filterEmpty = document.querySelector('[data-wings-admin-server-filter-empty]');
  const rows = Array.from(document.querySelectorAll('[data-wings-admin-server-row]'), (row) => ({
    row,
    searchText: Array.from(row.cells).slice(0, 3).map((cell) => cell.textContent).join(' ').toLowerCase(),
  }));
  if (!filterInput || !filterEmpty || !rows.length) return;

  filterInput.addEventListener('input', () => {
    const query = filterInput.value.trim().toLowerCase();
    let matches = 0;
    for (const {row, searchText} of rows) {
      const isMatch = searchText.includes(query);
      row.hidden = !isMatch;
      if (isMatch) matches++;
    }
    filterEmpty.hidden = matches > 0;
  });
}
