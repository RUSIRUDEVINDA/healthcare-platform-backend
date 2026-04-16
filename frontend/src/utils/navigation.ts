export function redirectToLogin() {
  if (typeof window === 'undefined') {
    return;
  }

  // Chrome's internal error pages (chrome-error://...) are not safe redirect origins.
  if (window.location.protocol !== 'http:' && window.location.protocol !== 'https:') {
    return;
  }

  const target = new URL('/auth/login', window.location.origin).toString();
  if (window.location.href === target) {
    return;
  }

  window.location.replace(target);
}
