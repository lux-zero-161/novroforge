const byteUnits = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB'];

const byteNumberFormat = new Intl.NumberFormat(navigator.language, {
  maximumFractionDigits: 1,
});

export function formatWingsBytes(bytes) {
  if (!Number.isFinite(bytes)) return '-';

  const sign = bytes < 0 ? '-' : '';
  let value = Math.abs(bytes);
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < byteUnits.length - 1) {
    value /= 1024;
    unitIndex++;
  }
  return `${sign}${byteNumberFormat.format(value)} ${byteUnits[unitIndex]}`;
}
