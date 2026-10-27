(function (global) {
  const moneyFormatter = new Intl.NumberFormat('es-CO', {
    style: 'currency',
    currency: 'COP',
    maximumFractionDigits: 0
  });

  function formatMoney(value) {
    const number = Number(value);
    return moneyFormatter.format(Number.isFinite(number) ? number : 0);
  }

  function parseNumericExpression(rawValue) {
    if (rawValue === null || rawValue === undefined || rawValue === '') {
      return 0;
    }

    const cleaned = String(rawValue)
        .replace(/\s+/g, '')
        .replace(/[^0-9+\-.,]/g, '')
        .replace(/\./g, '')
        .replace(/,(?=.*,)/g, '')
        .replace(/,/g, '.')
        .replace(/(^|[+\-])0+(?=\d)/g, '$1')
        .replace(/[+\-]+$/, '');

    if (cleaned === '' || cleaned === '+' || cleaned === '-' || cleaned === '+-' || cleaned === '-+') {
      return 0;
    }

    const terms = cleaned.match(/[+-]?\d*\.?\d+/g);
    if (!terms) {
      return 0;
    }

    const result = terms.reduce((sum, term) => sum + Number.parseFloat(term), 0);
    return Number.isFinite(result) ? result : 0;
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

  function getInputNumber(input) {
    return input ? parseNumericExpression(input.value) : 0;
  }

  function sanitizeAmountInput(input) {
    if (!input) {
      return;
    }
    input.value = input.value.replace(/[^0-9+\-.,]/g, '');
  }

  function formatAmountInput(input) {
    if (!input) {
      return;
    }
    input.value = formatNumberInput(getInputNumber(input));
  }

  function pad(value) {
    return String(value).padStart(2, '0');
  }

  function toLocalDateTimeInput(date) {
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
  }

  function toLocalDateKey(date) {
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
  }

  function dateFromKey(key) {
    const [year, month, day] = String(key).split('-').map(Number);
    if (!year || !month || !day) {
      return null;
    }
    return new Date(year, month - 1, day);
  }

  function formatDateKey(key) {
    const date = dateFromKey(key);
    if (!date) {
      return key || 'Sin fecha';
    }
    return date.toLocaleDateString('es-CO', { day: 'numeric', month: 'long', year: 'numeric' });
  }

  function formatRangeLabel(fromKey, toKey) {
    if (!fromKey && !toKey) {
      return 'Sin rango de fechas';
    }
    if (fromKey === toKey || !toKey) {
      return formatDateKey(fromKey);
    }
    return `Del ${formatDateKey(fromKey)} al ${formatDateKey(toKey)}`;
  }

  function formatDateTime(value) {
    if (!value) {
      return '';
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return String(value);
    }
    return date.toLocaleString('es-CO', {
      day: '2-digit',
      month: 'long',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function browserTimezone() {
    try {
      return Intl.DateTimeFormat().resolvedOptions().timeZone || '';
    } catch (error) {
      return '';
    }
  }

  async function fetchJSON(url, options) {
    const response = await fetch(url, options);
    if (response.status === 401) {
      window.location.href = '/login';
      throw new Error('La sesión venció');
    }

    let payload = null;
    try {
      payload = await response.json();
    } catch (error) {
      payload = null;
    }

    if (!response.ok) {
      const message = payload && payload.error ? payload.error : 'No se pudo completar la operación';
      throw new Error(message);
    }
    return payload;
  }

  global.ProfitCommon = {
    formatMoney,
    parseNumericExpression,
    formatNumberInput,
    getInputNumber,
    sanitizeAmountInput,
    formatAmountInput,
    toLocalDateTimeInput,
    toLocalDateKey,
    dateFromKey,
    formatDateKey,
    formatRangeLabel,
    formatDateTime,
    browserTimezone,
    fetchJSON
  };
})(window);
