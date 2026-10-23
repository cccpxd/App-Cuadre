const balanceBox = document.getElementById('avicolaBalance');
const historyBox = document.getElementById('avicolaHistory');
const dateFilter = document.getElementById('avicolaDateFilter');

function formatMoney(value) {
  return new Intl.NumberFormat('es-CO', {
    style: 'currency',
    currency: 'COP',
    maximumFractionDigits: 0
  }).format(Number(value || 0));
}

function formatDate(dateString) {
  if (!dateString) {
    return '';
  }

  const date = new Date(dateString);
  if (Number.isNaN(date.getTime())) {
    return dateString;
  }

  return date.toLocaleString('es-CO', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
}

async function loadAvicolaData() {
  if (!historyBox || !balanceBox) {
    return;
  }

  const params = new URLSearchParams();
  const selectedDate = dateFilter ? dateFilter.value : '';
  if (selectedDate) {
    params.set('date', selectedDate);
  }

  const response = await fetch(`/api/avicola?${params.toString()}`);
  const payload = await response.json();
  const movements = Array.isArray(payload.movements) ? payload.movements : [];

  historyBox.innerHTML = '';
  movements.forEach(item => {
    const signedAmount = item.movement_type === 'egreso' ? -(Number(item.amount) || 0) : Number(item.amount) || 0;
    const row = document.createElement('div');
    row.className = 'history-item';
    row.innerHTML = `
      <strong>${item.movement_type}</strong>
      <span>${item.description}</span>
      <small>${formatDate(item.entry_date)}</small>
      <span>${formatMoney(signedAmount)}</span>
    `;
    historyBox.appendChild(row);
  });

  if (movements.length === 0) {
    historyBox.innerHTML = '<div class="empty-state-panel"><p>No hay movimientos para esta fecha.</p></div>';
  }

  balanceBox.textContent = formatMoney(Number(payload.balance || 0));
}

const saveButton = document.getElementById('saveAvicolaMovement');
if (saveButton) {
  saveButton.addEventListener('click', async () => {
    const payload = {
      description: document.getElementById('movementDescription').value,
      amount: Number(document.getElementById('movementAmount').value || 0),
      movement_type: document.getElementById('movementType').value,
      entry_date: new Date().toISOString()
    };

    const response = await fetch('/api/avicola', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    if (response.ok) {
      document.getElementById('movementDescription').value = '';
      document.getElementById('movementAmount').value = 0;
      loadAvicolaData();
    } else {
      const data = await response.json();
      alert(data.error || 'No se pudo guardar el movimiento');
    }
  });
}

if (dateFilter) {
  dateFilter.addEventListener('change', loadAvicolaData);
}

if (historyBox && balanceBox) {
  const today = new Date().toISOString().slice(0, 10);
  if (dateFilter) {
    dateFilter.value = today;
  }
  loadAvicolaData();
}
