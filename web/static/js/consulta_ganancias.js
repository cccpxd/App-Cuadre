(function () {
  const common = window.ProfitCommon;
  const recordsBox = document.getElementById('profitRecords');

  if (!common || !recordsBox) {
    return;
  }

  const summaryBox = document.getElementById('profitSummary');
  const rangeLabel = document.getElementById('profitRangeLabel');
  const fromInput = document.getElementById('profitFrom');
  const toInput = document.getElementById('profitTo');
  const presetButtons = Array.from(document.querySelectorAll('.profit-preset'));

  const sectionLabels = {
    caja_principal: 'Caja principal',
    caja_auxiliar: 'Caja auxiliar'
  };
  const sectionOrder = ['caja_principal', 'caja_auxiliar'];

  let activePeriod = '';
  let manualRange = null;
  let currentDay = common.toLocalDateKey(new Date());

  function setActivePreset(period) {
    presetButtons.forEach(button => {
      button.classList.toggle('active', button.dataset.period === period);
    });
  }

  function buildMetricBox(label, value, modifier) {
    const box = document.createElement('div');
    box.className = 'metric-box profit-metric' + (modifier ? ` ${modifier}` : '');

    const title = document.createElement('span');
    title.textContent = label;

    const amount = document.createElement('strong');
    amount.textContent = value;

    box.appendChild(title);
    box.appendChild(amount);
    return box;
  }

  function renderSummary(payload) {
    if (!summaryBox) {
      return;
    }

    summaryBox.innerHTML = '';
    const totals = payload.totals || {};
    const records = Array.isArray(payload.records) ? payload.records : [];

    summaryBox.appendChild(buildMetricBox('Ganancia', common.formatMoney(totals.ganancia), 'positive'));
    summaryBox.appendChild(buildMetricBox('Caja principal', common.formatMoney(totals.caja_principal)));
    summaryBox.appendChild(buildMetricBox('Caja auxiliar', common.formatMoney(totals.caja_auxiliar)));
    summaryBox.appendChild(buildMetricBox(
        'Registros',
        records.length === 1 ? '1' : String(records.length)
    ));
  }

  function buildDetail(record) {
    const detail = document.createElement('div');
    detail.className = 'profit-record-detail';
    detail.hidden = true;

    const lines = Array.isArray(record.lines) ? record.lines : [];
    if (lines.length === 0) {
      const empty = document.createElement('p');
      empty.className = 'profit-empty-detail';
      empty.textContent = 'Este registro no tiene campos guardados.';
      detail.appendChild(empty);
      return detail;
    }

    sectionOrder.forEach(section => {
      const sectionLines = lines
          .filter(line => line.section === section)
          .sort((a, b) => Number(a.order_index) - Number(b.order_index));

      if (sectionLines.length === 0) {
        return;
      }

      const group = document.createElement('div');
      group.className = 'profit-detail-group';

      const heading = document.createElement('h4');
      heading.textContent = sectionLabels[section] || section;
      group.appendChild(heading);

      sectionLines.forEach(line => {
        const row = document.createElement('div');
        row.className = 'profit-detail-line';

        const concept = document.createElement('span');
        concept.textContent = line.concept_name;

        const amount = document.createElement('strong');
        amount.textContent = common.formatMoney(line.amount);

        row.appendChild(concept);
        row.appendChild(amount);
        group.appendChild(row);
      });

      detail.appendChild(group);
    });

    return detail;
  }

  function buildRecordCard(record) {
    const card = document.createElement('article');
    card.className = 'history-card profit-record-card';

    const header = document.createElement('div');
    header.className = 'history-card-header';

    const titleBox = document.createElement('div');
    const kicker = document.createElement('span');
    kicker.className = 'history-number';
    kicker.textContent = 'Ganancia avícola';

    const title = document.createElement('h3');
    title.textContent = common.formatDateTime(record.recorded_at);

    titleBox.appendChild(kicker);
    titleBox.appendChild(title);

    if (record.notes) {
      const notes = document.createElement('p');
      notes.className = 'history-author';
      notes.textContent = record.notes;
      titleBox.appendChild(notes);
    }

    header.appendChild(titleBox);

    const metrics = document.createElement('div');
    metrics.className = 'history-metrics';
    const totals = record.totals || {};
    metrics.appendChild(buildMetricBox('Ganancia', common.formatMoney(totals.ganancia)));
    metrics.appendChild(buildMetricBox('Caja principal', common.formatMoney(totals.caja_principal)));
    metrics.appendChild(buildMetricBox('Caja auxiliar', common.formatMoney(totals.caja_auxiliar)));

    const detail = buildDetail(record);

    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'button secondary profit-detail-button';
    button.textContent = 'Ver detalle';
    button.addEventListener('click', () => {
      const willShow = detail.hidden;
      detail.hidden = !willShow;
      button.textContent = willShow ? 'Ocultar detalle' : 'Ver detalle';
    });

    card.appendChild(header);
    card.appendChild(metrics);
    card.appendChild(button);
    card.appendChild(detail);
    return card;
  }

  function renderRecords(payload) {
    const records = Array.isArray(payload.records) ? payload.records : [];
    recordsBox.innerHTML = '';

    if (records.length === 0) {
      const empty = document.createElement('div');
      empty.className = 'history-empty';
      empty.textContent = 'No hay ganancias registradas en este rango.';
      recordsBox.appendChild(empty);
      return;
    }

    records.forEach(record => {
      recordsBox.appendChild(buildRecordCard(record));
    });
  }

  function render(payload) {
    const range = payload.range || {};
    rangeLabel.textContent = common.formatRangeLabel(range.from_date, range.to_date);

    if (fromInput && range.from_date) {
      fromInput.value = range.from_date;
    }
    if (toInput && range.to_date) {
      toInput.value = range.to_date;
    }

    renderSummary(payload);
    renderRecords(payload);
  }

  function renderError(message) {
    if (rangeLabel) {
      rangeLabel.textContent = 'No se pudo cargar la consulta';
    }
    if (summaryBox) {
      summaryBox.innerHTML = '';
    }
    recordsBox.innerHTML = '';
    const empty = document.createElement('div');
    empty.className = 'history-empty';
    empty.textContent = message || 'No se pudo cargar la consulta. Intenta de nuevo.';
    recordsBox.appendChild(empty);
  }

  async function loadQuery(url) {
    if (rangeLabel) {
      rangeLabel.textContent = 'Cargando consulta...';
    }

    try {
      const payload = await common.fetchJSON(url);
      render(payload);
    } catch (error) {
      renderError(error.message);
    }
  }

  function loadPeriod(period) {
    activePeriod = period;
    manualRange = null;
    setActivePreset(period);

    const params = new URLSearchParams();
    params.set('period', period);
    params.set('reference_date', common.toLocalDateKey(new Date()));
    const timezone = common.browserTimezone();
    if (timezone) {
      params.set('timezone', timezone);
    }

    loadQuery(`/api/ganancias-avicola?${params.toString()}`);
  }

  function loadRange(from, to) {
    activePeriod = '';
    manualRange = { from, to };
    setActivePreset('');

    const params = new URLSearchParams();
    params.set('from', from);
    params.set('to', to);
    const timezone = common.browserTimezone();
    if (timezone) {
      params.set('timezone', timezone);
    }

    loadQuery(`/api/ganancias-avicola?${params.toString()}`);
  }

  presetButtons.forEach(button => {
    button.addEventListener('click', () => {
      if (button.dataset.period) {
        loadPeriod(button.dataset.period);
      }
    });
  });

  const applyButton = document.getElementById('applyProfitRange');
  if (applyButton) {
    applyButton.addEventListener('click', () => {
      const from = fromInput ? fromInput.value : '';
      const to = toInput ? toInput.value : '';

      if (!from || !to) {
        window.alert('Selecciona la fecha inicial y la final.');
        return;
      }
      if (to < from) {
        window.alert('La fecha final no puede ser anterior a la inicial.');
        return;
      }

      loadRange(from, to);
    });
  }

  function refreshIfDayChanged() {
    const today = common.toLocalDateKey(new Date());
    if (today === currentDay) {
      return;
    }
    currentDay = today;

    if (activePeriod) {
      loadPeriod(activePeriod);
    } else if (manualRange) {
      loadRange(manualRange.from, manualRange.to);
    }
  }

  window.setInterval(refreshIfDayChanged, 30000);
  document.addEventListener('visibilitychange', () => {
    if (!document.hidden) {
      refreshIfDayChanged();
    }
  });

  const requestedPeriod = new URLSearchParams(window.location.search).get('period');
  const availablePeriods = presetButtons.map(button => button.dataset.period);
  const initialPeriod = availablePeriods.includes(requestedPeriod) ? requestedPeriod : 'semanal';
  loadPeriod(initialPeriod);
})();
