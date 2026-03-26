const APP_BASE = window.APP_BASE || '';
const API_BASE = APP_BASE + '/api/v1';

const App = {
  token: localStorage.getItem('access_token'),
  refreshToken: localStorage.getItem('refresh_token'),
  user: JSON.parse(localStorage.getItem('user') || 'null'),

  isLoggedIn() {
    return !!this.token;
  },

  setAuth(tokens, user) {
    this.token = tokens.access_token;
    this.refreshToken = tokens.refresh_token;
    this.user = user;
    localStorage.setItem('access_token', tokens.access_token);
    localStorage.setItem('refresh_token', tokens.refresh_token);
    localStorage.setItem('user', JSON.stringify(user));
  },

  clearAuth() {
    this.token = null;
    this.refreshToken = null;
    this.user = null;
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user');
  },

  async apiForm(method, path, formData) {
    const opts = { method, body: formData, headers: {} };
    if (this.token) {
      opts.headers['Authorization'] = 'Bearer ' + this.token;
    }

    let res = await fetch(API_BASE + path, opts);

    if (res.status === 401 && this.refreshToken) {
      const refreshed = await this.tryRefresh();
      if (refreshed) {
        if (this.token) {
          opts.headers['Authorization'] = 'Bearer ' + this.token;
        }
        res = await fetch(API_BASE + path, opts);
      } else {
        this.clearAuth();
        window.location.href = APP_BASE + '/';
        return null;
      }
    }

    const data = await res.json();
    if (!res.ok) {
      throw { status: res.status, ...data };
    }
    return data;
  },

  async api(method, path, body = null) {
    const opts = {
      method,
      headers: { 'Content-Type': 'application/json' },
    };
    if (this.token) {
      opts.headers['Authorization'] = 'Bearer ' + this.token;
    }
    if (body) {
      opts.body = JSON.stringify(body);
    }

    let res = await fetch(API_BASE + path, opts);

    if (res.status === 401 && this.refreshToken) {
      const refreshed = await this.tryRefresh();
      if (refreshed) {
        opts.headers['Authorization'] = 'Bearer ' + this.token;
        res = await fetch(API_BASE + path, opts);
      } else {
        this.clearAuth();
        window.location.href = APP_BASE + '/';
        return null;
      }
    }

    const data = await res.json();
    if (!res.ok) {
      throw { status: res.status, ...data };
    }
    return data;
  },

  async tryRefresh() {
    try {
      const res = await fetch(API_BASE + '/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: this.refreshToken }),
      });
      if (!res.ok) return false;
      const data = await res.json();
      this.token = data.data.tokens.access_token;
      this.refreshToken = data.data.tokens.refresh_token;
      localStorage.setItem('access_token', this.token);
      localStorage.setItem('refresh_token', this.refreshToken);
      return true;
    } catch {
      return false;
    }
  },

  requireAuth() {
    if (!this.isLoggedIn()) {
      window.location.href = APP_BASE + '/';
      return false;
    }
    return true;
  },

  async logout() {
    try {
      await this.api('POST', '/auth/logout', { refresh_token: this.refreshToken });
    } catch (e) { /* ignore */ }
    this.clearAuth();
    window.location.href = APP_BASE + '/';
  },

  showToast(message, type = 'success') {
    let container = document.querySelector('.toast-container');
    if (!container) {
      container = document.createElement('div');
      container.className = 'toast-container';
      document.body.appendChild(container);
    }

    const colors = {
      success: 'bg-success',
      error: 'bg-danger',
      warning: 'bg-warning text-dark',
      info: 'bg-info text-dark'
    };

    const id = 'toast-' + Date.now();
    const html = `
      <div id="${id}" class="toast align-items-center text-white ${colors[type] || colors.success} border-0"
           role="alert" data-bs-autohide="true" data-bs-delay="4000">
        <div class="d-flex">
          <div class="toast-body">${message}</div>
          <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
        </div>
      </div>`;
    container.insertAdjacentHTML('beforeend', html);
    const el = document.getElementById(id);
    new bootstrap.Toast(el).show();
    el.addEventListener('hidden.bs.toast', () => el.remove());
  },

  initSidebar() {
    const user = this.user;
    if (!user) return;

    document.querySelectorAll('.user-name').forEach(el => el.textContent = user.full_name || user.email);
    document.querySelectorAll('.user-role').forEach(el => el.textContent = user.role);
    document.querySelectorAll('.user-facility').forEach(el => {
      if (user.role === 'district_biostatistician' && user.district) {
        el.textContent = 'District: ' + user.district;
      } else {
        el.textContent = user.facility ? user.facility.name : 'All Facilities';
      }
    });

    const role = user.role;
    const adminNavRoles = ['superadmin', 'facility_admin', 'district_biostatistician'];
    if (!adminNavRoles.includes(role)) {
      document.querySelectorAll('.admin-only').forEach(el => el.style.display = 'none');
    }
    const auditNavRoles = ['superadmin', 'facility_admin', 'district_biostatistician'];
    if (!auditNavRoles.includes(role)) {
      document.querySelectorAll('.facility-admin-audit').forEach(el => el.style.display = 'none');
    }
    if (role !== 'superadmin') {
      document.querySelectorAll('.superadmin-only').forEach(el => el.style.display = 'none');
    }

    const path = window.location.pathname;
    document.querySelectorAll('.sidebar .nav-link').forEach(link => {
      const href = link.getAttribute('href');
      if (href === path || (href !== '/' && path.startsWith(href))) {
        link.classList.add('active');
      }
    });
  },

  formatDate(dateStr) {
    if (!dateStr) return '';
    const d = new Date(dateStr);
    return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' });
  }
};
