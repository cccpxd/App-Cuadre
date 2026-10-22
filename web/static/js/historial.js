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
    item.innerHTML = `
      <strong>${record.record_date}</strong><br>
      Base: ${new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP' }).format(Number(record.base_value || 0))}<br>
      Total: ${new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP' }).format(Number(record.total_value || 0))}<br>
      Diferencia: ${new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP' }).format(Number(record.difference || 0))}
    `;
    container.appendChild(item);
  });
}

if (document.getElementById('cuadreHistory')) {
  loadHistory();
}
