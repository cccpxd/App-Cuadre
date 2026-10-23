const moneyFormatter = new Intl.NumberFormat('es-CO', {
  style: 'currency',
  currency: 'COP',
  maximumFractionDigits: 0
});

function formatDate(dateString) {
  if (!dateString) {
    return 'Sin fecha';
  }

  const date = new Date(dateString);
  if (Number.isNaN(date.getTime())) {
    return 'Sin fecha';
  }

  return new Intl.DateTimeFormat('es-CO', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  }).format(date);
}

function getDateFilterValue() {
  const dateInput = document.getElementById('historyDateFilter');
  return dateInput ? dateInput.value : '';
}

async function loadHistory(dateFilter = '') {
  const container = document.getElementById('cuadreHistory');
  if (!container) {
    return;
  }

  const url = dateFilter ? `/api/cuadres?date=${encodeURIComponent(dateFilter)}` : '/api/cuadres';
  const response = await fetch(url);
  const records = await response.json();

  container.innerHTML = '';
  if (!records || records.length === 0) {
    container.innerHTML = '<div class="history-empty">No hay cuadres registrados para esta fecha.</div>';
    return;
  }

  records.forEach((record, index) => {
    const item = document.createElement('article');
    item.className = 'history-card';

    const diff = Number(record.difference || 0);
    const diffClass = diff >= 0 ? 'positive' : 'negative';

    item.innerHTML = `
      <div class="history-card-header">
        <div>
          <span class="history-number">Cuadre #${index + 1}</span>
          <h3>${formatDate(record.record_date)}</h3>
        </div>
        <span class="status-badge ${diffClass}">${diff >= 0 ? 'Sobra' : 'Falta'}</span>
      </div>
      <div class="history-metrics">
        <div class="metric-box">
          <span>Base</span>
          <strong>${moneyFormatter.format(Number(record.base_value || 0))}</strong>
        </div>
        <div class="metric-box">
          <span>Total</span>
          <strong>${moneyFormatter.format(Number(record.total_value || 0))}</strong>
        </div>
        <div class="metric-box">
          <span>Diferencia</span>
          <strong>${moneyFormatter.format(Number(record.difference || 0))}</strong>
        </div>
      </div>
      <div class="history-note">${record.notes ? record.notes : 'Sin observaciones'}</div>
    `;
    container.appendChild(item);
  });
}

const filterButton = document.getElementById('applyHistoryFilter');
const clearButton = document.getElementById('clearHistoryFilter');

if (filterButton) {
  filterButton.addEventListener('click', () => {
    loadHistory(getDateFilterValue());
  });
}

if (clearButton) {
  clearButton.addEventListener('click', () => {
    const dateInput = document.getElementById('historyDateFilter');
    if (dateInput) {
      dateInput.value = '';
    }
    loadHistory();
  });
}

if (document.getElementById('cuadreHistory')) {
  loadHistory();
}
