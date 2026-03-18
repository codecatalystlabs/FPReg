document.addEventListener('DOMContentLoaded', async () => {
  if (!App.requireAuth()) return;
  App.initSidebar();

  const yearEl = document.getElementById('filterYear');
  const monthEl = document.getElementById('filterMonth');
  const facilityEl = document.getElementById('filterFacilityId');
  const btn = document.getElementById('btnLoadReport');
  const tbody = document.getElementById('fpReportBody');

  const now = new Date();
  yearEl.value = now.getFullYear();
  const mm = String(now.getMonth() + 1).padStart(2, '0');
  monthEl.value = mm;

  const METHODS = [
    { code: 'FP01', name: 'Combined Oral Contraceptives Pills (COCs)' },
    { code: 'FP02', name: 'Progesterone Only Pills (POP)' },
    { code: 'FP03', name: 'Emergency Contraceptive Pills (ECP)' },
    { code: 'FP04', name: 'Injectable DMPA Intramuscular (3 months)' },
    { code: 'FP05_PA', name: 'Injectable DMPA-SC (PA)' },
    { code: 'FP05_SI', name: 'Injectable DMPA-SC (SI)' },
    { code: 'FP06', name: '3-year Implant' },
    { code: 'FP07', name: '5-year Implant' },
    { code: 'FP08', name: 'IUD Copper-T' },
    { code: 'FP09', name: 'IUD Hormonal' },
    { code: 'FP10', name: 'FAM-SDM (Cycle Beads)' },
    { code: 'FP11', name: 'FAM-LAM' },
    { code: 'FP12', name: 'FAM Two Day Method' },
    { code: 'FP13', name: 'Female Condoms' },
    { code: 'FP14', name: 'Male Condoms' },
  ];

  const AGE_KEYS = ['BELOW_15', '16_19', '20_24', '25_49', '50_PLUS'];

  async function loadReport() {
    const year = parseInt(yearEl.value, 10);
    const month = monthEl.value;
    if (!year || !month) {
      App.showToast('Select year and month', 'warning');
      return;
    }
    const period = String(year) + month;
    const params = new URLSearchParams({ period });
    if (facilityEl.value.trim()) params.set('facility_id', facilityEl.value.trim());

    tbody.innerHTML = '<tr><td colspan="7" class="text-center py-4"><div class="spinner-border spinner-border-sm text-primary"></div></td></tr>';

    try {
      const res = await App.api('GET', '/reports/family-planning/monthly?' + params.toString());
      const data = res.data;
      if (!data || !data.facilities || data.facilities.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" class="text-center py-4 text-muted">No data for selected period.</td></tr>';
        return;
      }
      const facility = data.facilities[0]; // for now show first; future: facility filter
      const cells = facility.cells || [];

      const matrix = {};
      METHODS.forEach(m => {
        matrix[m.code] = {
          NEW: { BELOW_15: 0, '16_19': 0, '20_24': 0, '25_49': 0, '50_PLUS': 0 },
          REVISIT: { BELOW_15: 0, '16_19': 0, '20_24': 0, '25_49': 0, '50_PLUS': 0 },
        };
      });

      cells.forEach(c => {
        const code = c.method;
        if (!matrix[code]) return;
        if (!matrix[code][c.visit_type]) return;
        if (matrix[code][c.visit_type][c.age_group] == null) return;
        matrix[code][c.visit_type][c.age_group] += c.value;
      });

      const rows = [];
      METHODS.forEach(m => {
        ['NEW', 'REVISIT'].forEach(vt => {
          const row = matrix[m.code][vt];
          rows.push(`
            <tr>
              <td>${m.name}</td>
              <td>${vt === 'NEW' ? 'New users' : 'Revisits'}</td>
              ${AGE_KEYS.map(a => `<td class="text-end">${row[a] || 0}</td>`).join('')}
            </tr>
          `);
        });
      });

      tbody.innerHTML = rows.join('');
    } catch (e) {
      console.error(e);
      App.showToast('Failed to load report', 'error');
      tbody.innerHTML = '<tr><td colspan="7" class="text-center py-4 text-muted">Error loading report.</td></tr>';
    }
  }

  btn.addEventListener('click', loadReport);
});

