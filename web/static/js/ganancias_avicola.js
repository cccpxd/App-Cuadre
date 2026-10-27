(function () {
  const common = window.ProfitCommon;
  const saveButton = document.getElementById('saveProfit');

  if (!common || !saveButton) {
    return;
  }

  const recordedAtInput = document.getElementById('profitRecordedAt');
  const notesInput = document.getElementById('profitNotes');
  const lineTemplate = document.getElementById('profitLineTemplate');
  const sections = Array.from(document.querySelectorAll('.profit-section'));
  const defaultConcepts = new Map();

  function linesContainer(section) {
    return section.querySelector('[data-section-lines]');
  }

  function updateSectionTotal(section) {
    const total = Array.from(section.querySelectorAll('.profit-amount'))
        .reduce((sum, input) => sum + common.getInputNumber(input), 0);
    const code = section.dataset.section;
    const target = section.querySelector(`[data-section-total="${code}"]`);
    if (target) {
      target.textContent = common.formatMoney(total);
    }
    return total;
  }

  // La ganancia no se escribe: es la suma de los totales de las dos cajas.
  function updateAllTotals() {
    const gain = sections
        .map(updateSectionTotal)
        .reduce((sum, total) => sum + total, 0);
    const gainTarget = document.getElementById('profitGainTotal');
    if (gainTarget) {
      gainTarget.textContent = common.formatMoney(gain);
    }
  }

  function attachAmountEvents(line) {
    const amountInput = line.querySelector('.profit-amount');
    if (!amountInput) {
      return;
    }

    amountInput.addEventListener('input', () => {
      common.sanitizeAmountInput(amountInput);
      updateAllTotals();
    });

    amountInput.addEventListener('blur', () => {
      common.formatAmountInput(amountInput);
      updateAllTotals();
    });
  }

  function addLine(section) {
    if (!lineTemplate) {
      return;
    }

    const line = lineTemplate.content.firstElementChild.cloneNode(true);
    const conceptInput = line.querySelector('.profit-concept');
    linesContainer(section).appendChild(line);
    attachAmountEvents(line);

    if (conceptInput) {
      conceptInput.focus();
      conceptInput.select();
    }
    updateAllTotals();
  }

  function collectLines() {
    const lines = [];

    sections.forEach(section => {
      const sectionCode = section.dataset.section;
      const rows = Array.from(section.querySelectorAll('.profit-line'));

      rows.forEach((row, index) => {
        const conceptInput = row.querySelector('.profit-concept');
        const amountInput = row.querySelector('.profit-amount');
        const concept = conceptInput ? conceptInput.value.trim() : '';

        if (!concept) {
          throw new Error('Todos los campos deben tener un nombre.');
        }

        lines.push({
          section: sectionCode,
          concept_name: concept,
          amount: common.getInputNumber(amountInput),
          order_index: index
        });
      });
    });

    return lines;
  }

  function resetForm() {
    sections.forEach(section => {
      const container = linesContainer(section);
      container.querySelectorAll('.profit-line[data-default="false"]').forEach(row => row.remove());

      const defaultLine = container.querySelector('.profit-line[data-default="true"]');
      if (defaultLine) {
        const conceptInput = defaultLine.querySelector('.profit-concept');
        const amountInput = defaultLine.querySelector('.profit-amount');
        const originalConcept = defaultConcepts.get(section.dataset.section);

        if (conceptInput && originalConcept) {
          conceptInput.value = originalConcept;
        }
        if (amountInput) {
          amountInput.value = '0';
        }
      }
    });

    if (notesInput) {
      notesInput.value = '';
    }
    if (recordedAtInput) {
      recordedAtInput.value = common.toLocalDateTimeInput(new Date());
    }
    updateAllTotals();
  }

  function buildRecordedAt() {
    if (!recordedAtInput || !recordedAtInput.value) {
      return new Date().toISOString();
    }

    const date = new Date(recordedAtInput.value);
    if (Number.isNaN(date.getTime())) {
      throw new Error('La fecha y hora no son válidas.');
    }
    return date.toISOString();
  }

  sections.forEach(section => {
    const defaultLine = section.querySelector('.profit-line[data-default="true"]');
    if (defaultLine) {
      const conceptInput = defaultLine.querySelector('.profit-concept');
      defaultConcepts.set(section.dataset.section, conceptInput ? conceptInput.value : '');
    }

    section.querySelectorAll('.profit-line').forEach(attachAmountEvents);

    const addButton = section.querySelector('.add-profit-field');
    if (addButton) {
      addButton.addEventListener('click', () => addLine(section));
    }

    const container = linesContainer(section);
    if (container) {
      container.addEventListener('click', event => {
        const removeButton = event.target.closest('.remove-profit-line');
        if (!removeButton) {
          return;
        }
        if (window.confirm('¿Desea eliminar este campo?')) {
          const line = removeButton.closest('.profit-line');
          if (line) {
            line.remove();
            updateAllTotals();
          }
        }
      });
    }
  });

  if (recordedAtInput) {
    recordedAtInput.value = common.toLocalDateTimeInput(new Date());
  }

  saveButton.addEventListener('click', async () => {
    let payload;

    try {
      payload = {
        recorded_at: buildRecordedAt(),
        notes: notesInput ? notesInput.value.trim() : '',
        lines: collectLines()
      };
    } catch (error) {
      window.alert(error.message);
      return;
    }

    const previousText = saveButton.textContent;
    saveButton.disabled = true;
    saveButton.textContent = 'Guardando...';

    try {
      await common.fetchJSON('/api/ganancias-avicola', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      window.alert('Ganancia guardada correctamente.');
      resetForm();
    } catch (error) {
      window.alert(error.message || 'No se pudo guardar la ganancia.');
    } finally {
      saveButton.disabled = false;
      saveButton.textContent = previousText;
    }
  });

  const clearButton = document.getElementById('clearProfit');
  if (clearButton) {
    clearButton.addEventListener('click', resetForm);
  }

  updateAllTotals();
})();
