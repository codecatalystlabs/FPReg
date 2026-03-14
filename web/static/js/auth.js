document.addEventListener('DOMContentLoaded', () => {
  const form = document.getElementById('loginForm');
  if (!form) return;

  if (App.isLoggedIn()) {
    window.location.href = '/dashboard';
    return;
  }

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const email = document.getElementById('email').value.trim();
    const password = document.getElementById('password').value;
    const btn = form.querySelector('button[type="submit"]');
    const errorEl = document.getElementById('loginError');

    btn.disabled = true;
    btn.innerHTML = '<span class="spinner-border spinner-border-sm me-1"></span>Signing in...';
    errorEl.classList.add('d-none');

    try {
      const data = await App.api('POST', '/auth/login', { email, password });
      App.setAuth(data.data.tokens, data.data.user);
      window.location.href = '/dashboard';
    } catch (err) {
      errorEl.textContent = err.message || 'Invalid credentials';
      errorEl.classList.remove('d-none');
    } finally {
      btn.disabled = false;
      btn.innerHTML = '<i class="bi bi-box-arrow-in-right me-1"></i>Sign In';
    }
  });
});
