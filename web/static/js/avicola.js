const balanceBox = document.getElementById('avicolaBalance');
const historyBox = document.getElementById('avicolaHistory');

function formatMoney(value) {
  return new Intl.NumberFormat('es-CO', {
    style: 'currency',
    currency: 'COP',
    maximumFractionDigits: 0
  }).format(Number(value || 0));
}

async function loadAvicolaData() {
  if (!historyBox || !balanceBox) {
    return;
  }

  const response = await fetch('/api/avicola');
  const movements = await response.json();

  historyBox.innerHTML = '';
  movements.forEach(item => {
    const signedAmount = item.movement_type === 'egreso' ? -(Number(item.amount) || 0) : Number(item.amount) || 0;
    const row = document.createElement('div');
    row.className = 'history-item';
    row.innerHTML = `<strong>${item.movement_type}</strong> - ${item.description}<br><small>${item.entry_date}</small><br><span>${formatMoney(signedAmount)}</span>`;
    historyBox.appendChild(row);
  });

  const currentBalance = movements.length > 0 ? Number(movements[0].balance || 0) : 0;
  balanceBox.textContent = formatMoney(currentBalance);
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

if (historyBox && balanceBox) {
  loadAvicolaData();
}
