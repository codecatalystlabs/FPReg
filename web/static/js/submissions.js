document.addEventListener('DOMContentLoaded', async () => {
  if (!App.requireAuth()) return;
  App.initSidebar();

  let currentPage = 1;
  const perPage = 25;

  await loadSubmissions();

  document.getElementById('searchBtn')?.addEventListener('click', () => {
    currentPage = 1;
    loadSubmissions();
  });

  document.getElementById('searchInput')?.addEventListener('keyup', (e) => {
    if (e.key === 'Enter') {
      currentPage = 1;
      loadSubmissions();
    }
  });

  document.getElementById('clearFilters')?.addEventListener('click', () => {
    document.getElementById('searchInput').value = '';
    document.getElementById('filterDateFrom').value = '';
    document.getElementById('filterDateTo').value = '';
    document.getElementById('filterSex').value = '';
    currentPage = 1;
    loadSubmissions();
  });

  async function loadSubmissions() {
    const tbody = document.getElementById('submissionsBody');
    const paginationEl = document.getElementById('pagination');
    if (!tbody) return;

    tbody.innerHTML = '<tr><td colspan="10" class="text-center py-4"><div class="spinner-border text-primary"></div></td></tr>';

    const params = new URLSearchParams({
      page: currentPage,
      per_page: perPage,
    });

    const search = document.getElementById('searchInput')?.value?.trim();
    if (search) params.set('search', search);
    const dateFrom = document.getElementById('filterDateFrom')?.value;
    if (dateFrom) params.set('date_from', dateFrom);
    const dateTo = document.getElementById('filterDateTo')?.value;
    if (dateTo) params.set('date_to', dateTo);
    const sex = document.getElementById('filterSex')?.value;
    if (sex) params.set('sex', sex);

    try {
      const data = await App.api('GET', '/registrations?' + params.toString());
      const items = data.data || [];
      const meta = data.meta;

      if (items.length === 0) {
        tbody.innerHTML = '<tr><td colspan="10" class="text-center text-muted py-4">No submissions found</td></tr>';
        if (paginationEl) paginationEl.innerHTML = '';
        updateCount(0);
        return;
      }

      tbody.innerHTML = items.map(r => `
        <tr>
          <td>${r.serial_number}</td>
          <td><span class="fw-semibold">${r.client_number || '<span class="text-muted">Visitor</span>'}</span></td>
          <td>${r.surname} ${r.given_name}</td>
          <td>${r.sex}</td>
          <td>${r.age}</td>
          <td>${r.visit_date}</td>
          <td>${r.is_new_user ? '<span class="badge badge-new">New</span>' : '<span class="badge badge-revisit">Revisit</span>'}</td>
          <td>${r.hts_code || '–'}</td>
          <td>${r.phone_number || '–'}</td>
          <td>
            <a href="${APP_BASE}/submission/${r.id}" class="btn btn-sm btn-outline-primary">
              <i class="bi bi-eye"></i>
            </a>
          </td>
        </tr>`).join('');

      updateCount(meta.total);
      renderPagination(paginationEl, meta);
    } catch (err) {
      tbody.innerHTML = '<tr><td colspan="10" class="text-center text-danger py-4">Failed to load submissions</td></tr>';
    }
  }

  function updateCount(total) {
    const el = document.getElementById('totalCount');
    if (el) el.textContent = total;
  }

  function renderPagination(container, meta) {
    if (!container || !meta) return;
    const pages = meta.total_pages;
    if (pages <= 1) { container.innerHTML = ''; return; }

    let html = '<nav><ul class="pagination pagination-sm mb-0">';
    html += `<li class="page-item ${currentPage <= 1 ? 'disabled' : ''}">
      <a class="page-link" href="#" data-page="${currentPage - 1}">&laquo;</a></li>`;

    for (let i = 1; i <= pages && i <= 10; i++) {
      html += `<li class="page-item ${i === currentPage ? 'active' : ''}">
        <a class="page-link" href="#" data-page="${i}">${i}</a></li>`;
    }

    html += `<li class="page-item ${currentPage >= pages ? 'disabled' : ''}">
      <a class="page-link" href="#" data-page="${currentPage + 1}">&raquo;</a></li>`;
    html += '</ul></nav>';

    container.innerHTML = html;
    container.querySelectorAll('a[data-page]').forEach(a => {
      a.addEventListener('click', (e) => {
        e.preventDefault();
        const p = parseInt(a.dataset.page);
        if (p >= 1 && p <= pages) {
          currentPage = p;
          loadSubmissions();
        }
      });
    });
  }
});
