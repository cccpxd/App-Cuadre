let allRecords = [];
let selectedDay = '';

function formatMoney(value) {
  return new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP' }).format(Number(value || 0));
}

function parseRecordDate(record) {
  const raw = record && (record.record_date || record.created_at);
  if (!raw) {
    return null;
  }
  const date = new Date(raw);
  return Number.isNaN(date.getTime()) ? null : date;
}

function toDateKey(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function todayKey() {
  return toDateKey(new Date());
}

function dateFromKey(key) {
  const [year, month, day] = String(key).split('-').map(Number);
  if (!year || !month || !day) {
    return null;
  }
  return new Date(year, month - 1, day);
}

function formatFullDate(date) {
  return date.toLocaleDateString('es-CO', { day: '2-digit', month: 'long', year: 'numeric' });
}

function formatTime(date) {
  return date.toLocaleTimeString('es-CO', { hour: '2-digit', minute: '2-digit' });
}

function formatDayHeading(date) {
  if (!date) {
    return 'Cuadres sin fecha';
  }
  const label = date.toLocaleDateString('es-CO', { day: 'numeric', month: 'long' });
  return date.getFullYear() === new Date().getFullYear()
      ? `Cuadres del ${label}`
      : `Cuadres del ${label} de ${date.getFullYear()}`;
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

function buildCard(record, date, number) {
  const card = document.createElement('div');
  card.className = 'history-card';

  const header = document.createElement('div');
  header.className = 'history-card-header';

  const titleBox = document.createElement('div');
  const label = document.createElement('span');
  label.className = 'history-number';
  label.textContent = date
      ? `Cuadre #${number} · ${formatTime(date)}`
      : `Cuadre #${number}`;
  const title = document.createElement('h3');
  title.textContent = date ? formatFullDate(date) : 'Sin fecha';
  titleBox.appendChild(label);
  titleBox.appendChild(title);
  const author = document.createElement('p');
  author.className = 'history-author';
  author.textContent = `Hecho por: ${record.prepared_by || 'Sin registrar'}`;
  titleBox.appendChild(author);

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

function buildDaySection(key, date, records) {
  const section = document.createElement('section');
  section.className = 'history-day';

  const header = document.createElement('div');
  header.className = 'history-day-header';

  const heading = document.createElement('h3');
  heading.textContent = formatDayHeading(date);
  header.appendChild(heading);

  if (records.length > 0) {
    const count = document.createElement('span');
    count.className = 'history-day-count';
    count.textContent = records.length === 1 ? '1 cuadre' : `${records.length} cuadres`;
    header.appendChild(count);
  }

  section.appendChild(header);

  const cards = document.createElement('div');
  cards.className = 'history-day-cards';

  if (records.length === 0) {
    const empty = document.createElement('div');
    empty.className = 'history-empty';
    empty.textContent = key === todayKey()
        ? 'Todavía no hay cuadres registrados hoy.'
        : 'No hay cuadres para esa fecha.';
    cards.appendChild(empty);
  } else {
    // La numeración se reinicia en cada día: el primer cuadre del día es el #1.
    const chronological = [...records].sort(
        (a, b) => a.date - b.date || Number(a.record.id) - Number(b.record.id)
    );
    const numbers = new Map();
    chronological.forEach((item, index) => numbers.set(item.record.id, index + 1));

    // En pantalla se muestra primero el más reciente.
    [...chronological].reverse().forEach(item => {
      cards.appendChild(buildCard(item.record, item.date, numbers.get(item.record.id)));
    });
  }

  section.appendChild(cards);
  return section;
}

function renderHistory() {
  const container = document.getElementById('cuadreHistory');
  if (!container) {
    return;
  }

  const targetKey = selectedDay || todayKey();
  const recordsForDay = [];

  allRecords.forEach(record => {
    const date = parseRecordDate(record);
    if (!date) {
      return;
    }
    if (toDateKey(date) === targetKey) {
      recordsForDay.push({ record, date });
    }
  });

  container.innerHTML = '';
  container.appendChild(buildDaySection(targetKey, dateFromKey(targetKey), recordsForDay));
}

async function loadHistory(day) {
  const container = document.getElementById('cuadreHistory');
  if (!container) {
    return;
  }

  selectedDay = day || '';

  let records = [];
  try {
    const response = await fetch('/api/cuadres');
    if (!response.ok) {
      throw new Error('No se pudo cargar el historial');
    }
    records = await response.json();
  } catch (error) {
    container.innerHTML = '';
    const empty = document.createElement('div');
    empty.className = 'history-empty';
    empty.textContent = 'No se pudo cargar el historial. Intenta de nuevo.';
    container.appendChild(empty);
    return;
  }

  allRecords = Array.isArray(records) ? records : [];
  renderHistory();
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

  let currentDay = todayKey();
  const refreshForNewDay = () => {
    const day = todayKey();
    if (day === currentDay) {
      return;
    }
    currentDay = day;
    if (dateInput) {
      dateInput.value = '';
    }
    // Al pasar el día el historial se limpia y la guía muestra la fecha nueva.
    loadHistory('');
  };

  window.setInterval(refreshForNewDay, 30000);
  document.addEventListener('visibilitychange', () => {
    if (!document.hidden) {
      refreshForNewDay();
    }
  });

  loadHistory('');
}

if (document.getElementById('cuadreHistory')) {
  initHistory();
}
