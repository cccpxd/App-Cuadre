function formatMoney(value) {
  return new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP' }).format(Number(value || 0));
}

function formatDate(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleDateString('es-CO', { day: '2-digit', month: 'long', year: 'numeric', timeZone: 'UTC' });
}

function buildMetric(label, value) {
  const box = document.createElement('div');
  box.className = 'metric-box';

  const title = document.createElement('span');
  title.textContent = label;

  const amount = document.createElement('strong');
  amount.textContent = formatMoney(value);

  box.appendChild(title);
  box.appendChild(amount);
  return box;
}

function buildDetail(record) {
  const detail = document.createElement('div');
  detail.style.display = 'none';
  detail.style.marginTop = '12px';
  detail.style.paddingTop = '10px';
  detail.style.borderTop = '1px solid #e2e8f0';

  const entries = [...(record.entries || [])].sort((a, b) => a.order_index - b.order_index);

  if (entries.length === 0) {
    detail.textContent = 'Este cuadre no tiene líneas guardadas.';
    return detail;
  }

  entries.forEach(entry => {
    const line = document.createElement('div');
    line.style.display = 'flex';
    line.style.justifyContent = 'space-between';
    line.style.gap = '12px';
    line.style.padding = '4px 0';

    const name = document.createElement('span');
    name.textContent = entry.name;

    const value = document.createElement('strong');
    value.textContent = formatMoney(entry.value);

    line.appendChild(name);
    line.appendChild(value);
    detail.appendChild(line);
  });

  return detail;
}

function buildCard(record) {
  const card = document.createElement('div');
  card.className = 'history-card';

  const header = document.createElement('div');
  header.className = 'history-card-header';

  const titleBox = document.createElement('div');
  const number = document.createElement('span');
  number.className = 'history-number';
  number.textContent = `Cuadre #${record.id}`;
  const title = document.createElement('h3');
  title.textContent = formatDate(record.record_date);
  titleBox.appendChild(number);
  titleBox.appendChild(title);

  const difference = Number(record.difference || 0);
  const badge = document.createElement('span');
  badge.className = 'status-badge ' + (difference >= 0 ? 'positive' : 'negative');
  badge.textContent = difference > 0 ? 'Sobrante' : difference < 0 ? 'Faltante' : 'Cuadrado';

  header.appendChild(titleBox);
  header.appendChild(badge);

  const metrics = document.createElement('div');
  metrics.className = 'history-metrics';
  metrics.appendChild(buildMetric('Base', record.base_value));
  metrics.appendChild(buildMetric('Total', record.total_value));
  metrics.appendChild(buildMetric('Diferencia', record.difference));

  card.appendChild(header);
  card.appendChild(metrics);

  if (record.notes) {
    const note = document.createElement('div');
    note.className = 'history-note';
    note.textContent = record.notes;
    card.appendChild(note);
  }

  const detail = buildDetail(record);

  const button = document.createElement('button');
  button.type = 'button';
  button.className = 'button secondary';
  button.style.marginTop = '12px';
  button.textContent = 'Ver detalle';
  button.addEventListener('click', () => {
    const isHidden = detail.style.display === 'none';
    detail.style.display = isHidden ? 'block' : 'none';
    button.textContent = isHidden ? 'Ocultar detalle' : 'Ver detalle';
  });

  card.appendChild(button);
  card.appendChild(detail);
  return card;
}

async function loadHistory(date) {
  const container = document.getElementById('cuadreHistory');
  if (!container) {
    return;
  }

  const url = date ? `/api/cuadres?date=${encodeURIComponent(date)}` : '/api/cuadres';
  const response = await fetch(url);
  const records = await response.json();

  container.innerHTML = '';
  if (!records || records.length === 0) {
    const empty = document.createElement('div');
    empty.className = 'history-empty';
    empty.textContent = date ? 'No hay cuadres para esa fecha.' : 'No hay cuadres registrados.';
    container.appendChild(empty);
    return;
  }

  records.forEach(record => container.appendChild(buildCard(record)));
}

function initHistory() {
  const dateInput = document.getElementById('historyDateFilter');
  const applyButton = document.getElementById('applyHistoryFilter');
  const clearButton = document.getElementById('clearHistoryFilter');

  if (applyButton && dateInput) {
    applyButton.addEventListener('click', () => loadHistory(dateInput.value));
  }

  if (clearButton && dateInput) {
    clearButton.addEventListener('click', () => {
      dateInput.value = '';
      loadHistory('');
    });
  }

  loadHistory('');
}

if (document.getElementById('cuadreHistory')) {
  initHistory();
}