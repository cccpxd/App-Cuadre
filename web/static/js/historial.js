function formatMoney(value) {
  return new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP' }).format(Number(value || 0));
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

async function loadHistory() {
  const container = document.getElementById('cuadreHistory');
  if (!container) {
    return;
  }

  const response = await fetch('/api/cuadres');
  const records = await response.json();

  container.innerHTML = '';
  if (!records || records.length === 0) {
    container.innerHTML = '<div class="history-item">No hay cuadres registrados.</div>';
    return;
  }

  records.forEach(record => {
    const item = document.createElement('div');
    item.className = 'history-item';

    const summary = document.createElement('div');
    summary.innerHTML = `
      <strong>${record.record_date}</strong><br>
      Base: ${formatMoney(record.base_value)}<br>
      Total: ${formatMoney(record.total_value)}<br>
      Diferencia: ${formatMoney(record.difference)}
    `;

    const detail = buildDetail(record);

    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'button secondary';
    button.style.marginTop = '10px';
    button.textContent = 'Ver detalle';
    button.addEventListener('click', () => {
      const isHidden = detail.style.display === 'none';
      detail.style.display = isHidden ? 'block' : 'none';
      button.textContent = isHidden ? 'Ocultar detalle' : 'Ver detalle';
    });

    item.appendChild(summary);
    item.appendChild(button);
    item.appendChild(detail);
    container.appendChild(item);
  });
}

if (document.getElementById('cuadreHistory')) {
  loadHistory();
}