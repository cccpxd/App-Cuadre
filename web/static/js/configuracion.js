async function loadConfig() {
  const companyName = document.getElementById('companyName');
  const defaultCurrency = document.getElementById('defaultCurrency');
  const showAvicolaDashboard = document.getElementById('showAvicolaDashboard');
  const allowCustomFields = document.getElementById('allowCustomFields');

  if (!companyName || !defaultCurrency || !showAvicolaDashboard || !allowCustomFields) {
    return;
  }

  const response = await fetch('/api/configuracion');
  const config = await response.json();

  companyName.value = config.company_name || 'Correspondencia';
  defaultCurrency.value = config.default_currency || 'COP';
  showAvicolaDashboard.checked = config.show_avicola_dashboard ?? true;
  allowCustomFields.checked = config.allow_custom_fields ?? true;
}

const saveConfigButton = document.getElementById('saveConfig');
if (saveConfigButton) {
  saveConfigButton.addEventListener('click', async () => {
    const payload = {
      company_name: document.getElementById('companyName').value,
      default_currency: document.getElementById('defaultCurrency').value,
      show_avicola_dashboard: document.getElementById('showAvicolaDashboard').checked,
      allow_custom_fields: document.getElementById('allowCustomFields').checked
    };

    const response = await fetch('/api/configuracion', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    if (response.ok) {
      alert('Configuración guardada');
    } else {
      const data = await response.json();
      alert(data.error || 'No se pudo guardar la configuración');
    }
  });
}

if (document.getElementById('companyName')) {
  loadConfig();
}
