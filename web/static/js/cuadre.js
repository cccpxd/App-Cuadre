const defaultRows = [
  'Bancolombia',
  'BEMovil',
  'Punto Red',
  'Giros',
  'PTM',
  'PTM 2',
  'Banco de Bogotá',
  'Avícola',
  'Monedas',
  'Efectivo'
];

const grid = document.getElementById('cash-grid');
const baseValueInput = document.getElementById('baseValue');
const totalValue = document.getElementById('totalValue');
const differenceValue = document.getElementById('differenceValue');

function parseNumericExpression(rawValue) {
  if (rawValue === null || rawValue === undefined || rawValue === '') {
    return 0;
  }

  const cleaned = String(rawValue)
    .replace(/\s+/g, '')
    .replace(/\./g, '')
    .replace(/,/g, '.')
    .replace(/[^0-9+\-]/g, '');

  if (cleaned === '' || cleaned === '+' || cleaned === '-' || cleaned === '+-' || cleaned === '-+') {
    return 0;
  }

  try {
    const result = Function(`"use strict"; return (${cleaned});`)();
    return Number.isFinite(result) ? result : 0;
  } catch {
    return 0;
  }
}

function formatNumberInput(value) {
  const numericValue = Number.parseFloat(value);
  if (!Number.isFinite(numericValue)) {
    return '0';
  }

  const absoluteValue = Math.abs(numericValue);
  const integerPart = Math.trunc(absoluteValue).toString();
  const decimals = String(numericValue).includes('.') ? String(numericValue).split('.')[1] : '';
  const formattedInteger = integerPart.replace(/\B(?=(\d{3})+(?!\d))/g, '.');

  if (!decimals) {
    return numericValue < 0 ? `-${formattedInteger}` : formattedInteger;
  }

  return `${numericValue < 0 ? '-' : ''}${formattedInteger},${decimals}`;
}

function formatMoney(value) {
  return new Intl.NumberFormat('es-CO', {
    style: 'currency',
    currency: 'COP',
    maximumFractionDigits: 0
  }).format(Number(value || 0));
}

function getInputValue(input) {
  if (!input) return 0;
  const value = input.value;
  return parseNumericExpression(value);
}

function setFormattedInputValue(input) {
  if (!input) return;
  const numericValue = getInputValue(input);
  input.value = formatNumberInput(numericValue);
}

function updateValues() {
  const rows = [...document.querySelectorAll('.cash-row')];
  const base = getInputValue(baseValueInput);
  const total = rows.reduce((sum, row) => {
    const valueInput = row.querySelector('.cash-input');
    return sum + getInputValue(valueInput);
  }, 0);
  const difference = total - base;

  totalValue.textContent = formatMoney(total);
  differenceValue.textContent = formatMoney(difference);
}

function addRow() {
  const row = document.createElement('div');
  row.className = 'cash-row';
  row.dataset.name = 'Nuevo campo';
  row.innerHTML = `
    <div class="cash-row-header">
      <button type="button" class="remove-row" title="Eliminar campo" aria-label="Eliminar campo">×</button>
    </div>
    <label class="cash-label">
      <span>Nombre del campo</span>
      <input type="text" value="Nuevo campo" class="cash-name-input" />
    </label>
    <label class="cash-label">
      <span>Valor</span>
      <input type="text" value="0" inputmode="numeric" class="cash-input" />
    </label>
  `;
  grid.appendChild(row);

  const nameInput = row.querySelector('.cash-name-input');
  const valueInput = row.querySelector('.cash-input');
  const removeButton = row.querySelector('.remove-row');

  nameInput.addEventListener('input', () => {
    row.dataset.name = nameInput.value || 'Nuevo campo';
  });

  valueInput.addEventListener('input', () => {
    const sanitized = valueInput.value.replace(/[^0-9+\-.,]/g, '');
    valueInput.value = sanitized;
    updateValues();
  });

  valueInput.addEventListener('blur', () => {
    setFormattedInputValue(valueInput);
    updateValues();
  });

  removeButton.addEventListener('click', () => {
    const shouldDelete = window.confirm('¿Desea eliminar este campo?');
    if (shouldDelete) {
      row.remove();
      updateValues();
    }
  });
  updateValues();
}

baseValueInput.addEventListener('input', () => {
  const sanitized = baseValueInput.value.replace(/[^0-9+\-.,]/g, '');
  baseValueInput.value = sanitized;
  updateValues();
});
baseValueInput.addEventListener('blur', () => {
  setFormattedInputValue(baseValueInput);
  updateValues();
});

document.querySelectorAll('.cash-input').forEach(input => {
  input.addEventListener('input', () => {
    const sanitized = input.value.replace(/[^0-9+\-.,]/g, '');
    input.value = sanitized;
    updateValues();
  });
  input.addEventListener('blur', () => {
    setFormattedInputValue(input);
    updateValues();
  });
});

document.getElementById('add-row').addEventListener('click', addRow);

document.getElementById('saveCuadre').addEventListener('click', async () => {
  const entries = [...document.querySelectorAll('.cash-row')].map((row, index) => ({
    name: row.dataset.name || row.querySelector('.cash-name-input')?.value || `Campo ${index + 1}`,
    value: getInputValue(row.querySelector('.cash-input')),
    is_base: false,
    order_index: index
  }));

  const rawDate = document.getElementById('recordDate').value;
  const payload = {
    record_date: rawDate ? new Date(`${rawDate}T12:00:00`).toISOString() : new Date().toISOString(),
    base_value: getInputValue(baseValueInput),
    notes: 'Cuadre generado desde la interfaz',
    entries,
    prepared_by: document.getElementById('preparedBy').value.trim(),
  };

  const response = await fetch('/api/cuadres', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });

  if (response.ok) {
    alert('Cuadre guardado exitosamente');
    window.location.href = '/historial';
  } else {
    const data = await response.json();
    alert(data.error || 'No se pudo guardar el cuadre');
  }
});

document.getElementById('clearCuadre').addEventListener('click', () => {
  document.querySelectorAll('.cash-row .cash-input').forEach(input => {
    input.value = '0';
  });
  baseValueInput.value = '0';
  updateValues();
});

updateValues();
