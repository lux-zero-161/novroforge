export function initWingsFiles() {
  const filterInput = document.querySelector('[data-wings-file-filter]');
  const filterEmpty = document.querySelector('[data-wings-file-filter-empty]');
  const rows = Array.from(document.querySelectorAll('[data-wings-file-row]'), (row) => ({
    row,
    searchText: row.textContent.toLowerCase(),
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
