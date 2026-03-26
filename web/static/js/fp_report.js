document.addEventListener('DOMContentLoaded', async () => {
  if (!App.requireAuth()) return;
  App.initSidebar();

  const yearEl = document.getElementById('filterYear');
  const monthEl = document.getElementById('filterMonth');
  const facilityEl = document.getElementById('filterFacilityId');
  const forceResyncEl = document.getElementById('dhis2ForceResync');
  const btn = document.getElementById('btnLoadReport');
  const btnPreview = document.getElementById('btnPreviewDHIS2');
  const btnSync = document.getElementById('btnSyncDHIS2');
  const tbody = document.getElementById('fpReportBody');
  const dhModalEl = document.getElementById('dhis2Modal');

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

  const AGE_KEYS = ['BELOW_15', '15_19', '20_24', '25_49', '50_PLUS'];

  function escapeHtml(s) {
    const d = document.createElement('div');
    d.textContent = s == null ? '' : String(s);
    return d.innerHTML;
  }

  function showDHIS2Modal(title, summaryHtml, jsonObj) {
    document.getElementById('dhis2ModalTitle').textContent = title;
    document.getElementById('dhis2ModalSummary').innerHTML = summaryHtml;
    const pre = document.getElementById('dhis2ModalBody');
    try {
      pre.textContent = jsonObj != null ? JSON.stringify(jsonObj, null, 2) : '';
    } catch {
      pre.textContent = String(jsonObj);
    }
    const m = bootstrap.Modal.getOrCreateInstance(dhModalEl);
    m.show();
  }

  /** Follows API pagination until all facilities are loaded (per_page capped at 10000 server-side). */
  async function fetchAllFacilities() {
    const all = [];
    let page = 1;
    const perPage = 2000;
    for (;;) {
      const res = await App.api('GET', `/facilities?page=${page}&per_page=${perPage}`);
      const batch = res.data || [];
      all.push(...batch);
      const meta = res.meta;
      const totalPages = meta && meta.total_pages != null ? meta.total_pages : 1;
      if (page >= totalPages || batch.length === 0) break;
      page += 1;
    }
    return all;
  }

  const userRole = App.user && App.user.role;
  const lockedFacilityRoles = ['facility_user', 'reviewer', 'facility_admin'];

  function canPostToDHIS2() {
    return ['superadmin', 'facility_admin', 'district_biostatistician'].includes(userRole);
  }

  function canPreviewDHIS2() {
    return userRole !== 'facility_user';
  }

  function applyDhis2ControlsVisibility() {
    const wrap = document.getElementById('fpReportDhis2Controls');
    if (!wrap) return;
    if (!canPreviewDHIS2()) {
      wrap.classList.add('d-none');
      return;
    }
    wrap.classList.remove('d-none');
    if (!canPostToDHIS2()) {
      if (btnSync) btnSync.classList.add('d-none');
      const fc = forceResyncEl && forceResyncEl.closest('.form-check');
      if (fc) fc.classList.add('d-none');
    }
  }

  async function loadFacilitiesDropdown() {
    const sel = facilityEl;
    const helpEl = document.getElementById('facilityFilterHelp');

    if (lockedFacilityRoles.includes(userRole) && App.user && App.user.facility_id) {
      let f = App.user.facility;
      if (!f) {
        try {
          const res = await App.api('GET', `/facilities/${App.user.facility_id}`);
          f = res.data;
        } catch (e) {
          console.error(e);
          App.showToast('Could not load your facility', 'error');
        }
      }
      sel.innerHTML = '';
      const opt = document.createElement('option');
      opt.value = App.user.facility_id;
      opt.textContent = f ? `${f.name} (${f.code})` : 'Your facility';
      sel.appendChild(opt);
      sel.value = App.user.facility_id;
      sel.disabled = true;
      if (helpEl) {
        helpEl.textContent =
          'Your assigned facility (read-only). The monthly table includes only this site.';
      }
      return;
    }

    try {
      const list = await fetchAllFacilities();
      const keep = sel.value;
      sel.disabled = false;
      sel.innerHTML = '<option value="">All allowed facilities (combined totals in table)</option>';
      list.forEach((f) => {
        const opt = document.createElement('option');
        opt.value = f.id;
        opt.textContent = `${f.name} (${f.code})`;
        sel.appendChild(opt);
      });
      const uid = App.user && App.user.facility_id;
      if (userRole !== 'superadmin' && uid && Array.from(sel.options).some((o) => o.value === uid)) {
        sel.value = uid;
      } else if (keep && Array.from(sel.options).some((o) => o.value === keep)) {
        sel.value = keep;
      }
      if (helpEl) {
        if (userRole === 'superadmin') {
          helpEl.textContent =
            'Leave as “all” for national totals, or choose one facility (required for DHIS2 sync).';
        } else if (userRole === 'district_biostatistician') {
          helpEl.textContent =
            'Leave as “all” for your whole district, or pick one facility. DHIS2 uses each facility’s org unit.';
        } else {
          helpEl.textContent = 'Choose a facility to filter the table.';
        }
      }
    } catch (e) {
      console.error(e);
      App.showToast('Could not load facility list', 'error');
    }
  }

  function currentPeriod() {
    const year = parseInt(yearEl.value, 10);
    const month = monthEl.value;
    if (!year || !month) return '';
    return String(year) + month;
  }

  function emptyMatrix() {
    const matrix = {};
    METHODS.forEach((m) => {
      matrix[m.code] = {
        NEW: { BELOW_15: 0, '15_19': 0, '20_24': 0, '25_49': 0, '50_PLUS': 0 },
        REVISIT: { BELOW_15: 0, '15_19': 0, '20_24': 0, '25_49': 0, '50_PLUS': 0 },
      };
    });
    return matrix;
  }

  function addCellsToMatrix(matrix, cells) {
    (cells || []).forEach((c) => {
      const code = c.method_code || c.method;
      const vt = c.visit_type;
      const ag = c.age_group;
      if (!code || !matrix[code]) return;
      if (!vt || !matrix[code][vt]) return;
      if (matrix[code][vt][ag] == null) return;
      matrix[code][vt][ag] += c.value || 0;
    });
  }

  async function loadReport() {
    const period = currentPeriod();
    if (!period) {
      App.showToast('Select year and month', 'warning');
      return;
    }
    const params = new URLSearchParams({ period });
    if (facilityEl.value.trim()) params.set('facility_id', facilityEl.value.trim());

    tbody.innerHTML = '<tr><td colspan="7" class="text-center py-4"><div class="spinner-border spinner-border-sm text-primary"></div></td></tr>';

    try {
      const res = await App.api('GET', '/reports/family-planning/monthly?' + params.toString());
      const data = res.data;
      const facilities = (data && data.facilities) || [];
      if (facilities.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" class="text-center py-4 text-muted">No data for selected period.</td></tr>';
        return;
      }

      const matrix = emptyMatrix();
      facilities.forEach((f) => addCellsToMatrix(matrix, f.cells));

      const rows = [];
      METHODS.forEach((m) => {
        ['NEW', 'REVISIT'].forEach((vt) => {
          const row = matrix[m.code][vt];
          rows.push(`
            <tr>
              <td>${m.name}</td>
              <td>${vt === 'NEW' ? 'New users' : 'Revisits'}</td>
              ${AGE_KEYS.map((a) => `<td class="text-end">${row[a] || 0}</td>`).join('')}
            </tr>
          `);
        });
      });

      tbody.innerHTML = rows.join('');
    } catch (e) {
      console.error(e);
      const msg = (e && e.message) || 'Failed to load report';
      App.showToast(msg, 'error');
      tbody.innerHTML = '<tr><td colspan="7" class="text-center py-4 text-muted">Error loading report.</td></tr>';
    }
  }

  async function previewDHIS2() {
    const period = currentPeriod();
    if (!period) {
      App.showToast('Select year and month', 'warning');
      return;
    }
    const params = new URLSearchParams({ period });
    if (facilityEl.value.trim()) params.set('facility_id', facilityEl.value.trim());
    try {
      const res = await App.api('GET', '/reports/family-planning/payload-preview?' + params.toString());
      const previews = res.data || [];
      const n = previews.reduce((acc, p) => acc + (p.data_values && p.data_values.length ? p.data_values.length : 0), 0);
      const missing = previews.reduce((acc, p) => acc + (p.missing_mappings && p.missing_mappings.length ? p.missing_mappings.length : 0), 0);
      const mergedMiss = {};
      previews.forEach((p) => {
        (p.missing_mapping_with_values || []).forEach((r) => {
          const k = r.local_indicator_key;
          mergedMiss[k] = (mergedMiss[k] || 0) + (r.value || 0);
        });
      });
      const missRows = Object.keys(mergedMiss)
        .sort()
        .map((k) => `<tr><td><code class="small">${escapeHtml(k)}</code></td><td class="text-end">${mergedMiss[k]}</td></tr>`)
        .join('');
      const missTable =
        missRows.length > 0
          ? `<p class="mb-1"><strong>Counts with no DHIS2 UIDs yet</strong></p><div class="table-responsive"><table class="table table-sm table-bordered"><thead><tr><th>local_indicator_key</th><th class="text-end">Count</th></tr></thead><tbody>${missRows}</tbody></table></div>`
          : '';
      const summary = [
        '<p><strong>Payload preview</strong></p>',
        '<ul class="small mb-0">',
        `<li>Facilities in preview: <strong>${previews.length}</strong></li>`,
        `<li>Data values ready to post: <strong>${n}</strong></li>`,
        `<li>Missing mapping keys (distinct): <strong>${missing}</strong></li>`,
        '</ul>',
        missTable,
      ].join('');
      showDHIS2Modal('DHIS2 payload preview', summary, previews);
      App.showToast(`Preview: ${n} data value(s) across ${previews.length} org unit(s).`, 'info');
    } catch (e) {
      console.error(e);
      const msg = (e && e.message) || 'Preview failed';
      showDHIS2Modal('DHIS2 preview error', `<p class="text-danger">${escapeHtml(msg)}</p>`, e);
      App.showToast(msg, 'error');
    }
  }

  function formatSyncSummaryHtml(o) {
    if (!o) return '<p class="text-muted">No outcome payload.</p>';
    const parts = [
      '<p><strong>Sync summary</strong></p>',
      '<ul class="small">',
      `<li>Facilities in preview: <strong>${o.preview_facility_count ?? 0}</strong></li>`,
      `<li>Data values built (before “already synced” filter): <strong>${o.total_data_values ?? 0}</strong></li>`,
      `<li>HTTP POST batches to DHIS2: <strong>${o.posted_batches ?? 0}</strong></li>`,
      `<li>Skipped — no data values / mapping: <strong>${o.skipped_no_data_values ?? 0}</strong></li>`,
      `<li>Skipped — all values already synced: <strong>${o.skipped_already_synced ?? 0}</strong></li>`,
      '</ul>',
    ];
    const msgs = o.detail_messages || [];
    if (msgs.length) {
      parts.push('<p class="mb-1"><strong>What happened</strong></p><ul class="small mb-0">');
      msgs.forEach((m) => parts.push(`<li>${escapeHtml(m)}</li>`));
      parts.push('</ul>');
    }
    const mm = o.missing_mapping_with_values || [];
    if (mm.length) {
      parts.push(
        '<p class="mb-1 mt-2"><strong>Aggregates without DHIS2 mapping (set UIDs in <code>dhis2_mapping_item</code>)</strong></p>',
        '<div class="table-responsive"><table class="table table-sm table-bordered mb-0"><thead><tr><th>local_indicator_key</th><th class="text-end">Count</th></tr></thead><tbody>',
      );
      mm.forEach((row) => {
        parts.push(
          `<tr><td><code class="small">${escapeHtml(row.local_indicator_key)}</code></td><td class="text-end">${row.value ?? 0}</td></tr>`,
        );
      });
      parts.push('</tbody></table></div>');
    }
    const logs = o.logs || [];
    if (logs.length) {
      parts.push('<p class="mb-0 mt-2 small text-muted">Full request/response details are in the JSON below (response_body per batch).</p>');
    }
    return parts.join('');
  }

  async function syncToDHIS2() {
    const period = currentPeriod();
    if (!period) {
      App.showToast('Select year and month', 'warning');
      return;
    }
    const fid = facilityEl.value.trim();
    const role = App.user && App.user.role;
    if (role === 'superadmin' && !fid) {
      App.showToast('Select a facility, then Sync to DHIS2 (superadmin).', 'warning');
      return;
    }
    const scopeNote = fid ? 'this facility' : 'all facilities in your scope';
    if (!confirm(`Post ${period} aggregated FP data to DHIS2 for ${scopeNote}?`)) return;

    try {
      const body = {
        period,
        facility_ids: fid ? [fid] : [],
        force: !!(forceResyncEl && forceResyncEl.checked),
      };
      const res = await App.api('POST', '/reports/family-planning/sync', body);
      const o = res.data;
      const logs = (o && o.logs) || [];
      const ok = logs.filter((l) => l.success).length;
      const bad = logs.length - ok;
      const summaryHtml = formatSyncSummaryHtml(o);
      showDHIS2Modal('DHIS2 sync result', summaryHtml, o);
      if (logs.length === 0) {
        App.showToast('Sync finished: no HTTP requests (see modal for reasons).', 'warning');
      } else {
        App.showToast(`DHIS2: ${ok} batch(es) ok, ${bad} failed.`, bad ? 'warning' : 'success');
      }
    } catch (e) {
      console.error(e);
      const msg = (e && e.message) || 'Sync failed';
      const extra = e.errors ? JSON.stringify(e.errors, null, 2) : '';
      showDHIS2Modal(
        'DHIS2 sync error',
        `<p class="text-danger">${escapeHtml(msg)}</p>${extra ? `<pre class="small">${escapeHtml(extra)}</pre>` : ''}`,
        e
      );
      App.showToast(msg, 'error');
    }
  }

  btn.addEventListener('click', loadReport);
  if (btnPreview) btnPreview.addEventListener('click', previewDHIS2);
  if (btnSync) btnSync.addEventListener('click', syncToDHIS2);

  await loadFacilitiesDropdown();
  applyDhis2ControlsVisibility();
  loadReport();
});
