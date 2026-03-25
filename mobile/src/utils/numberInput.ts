/**
 * Helpers for controlled numeric TextInputs (number-pad).
 * Avoids parseInt(x) || 0 and min-clamping on every keystroke (breaks multi-digit entry).
 */

/** Digits only; empty → 0; clamp to [0, max]. */
export function parseQtyWhileTyping(raw: string, max: number): number {
  const digits = raw.replace(/\D/g, '');
  if (digits === '') return 0;
  let n = parseInt(digits, 10);
  if (!Number.isFinite(n)) return 0;
  if (n < 0) return 0;
  if (n > max) n = max;
  return n;
}

/** Digits only; empty → 0; no upper bound (e.g. female condoms). */
export function parseNonNegWhileTyping(raw: string): number {
  const digits = raw.replace(/\D/g, '');
  if (digits === '') return 0;
  const n = parseInt(digits, 10);
  if (!Number.isFinite(n) || n < 0) return 0;
  return n;
}

/**
 * Age while typing: digits only, cap at max only.
 * Do not enforce min on each keystroke (so "25" can be typed as 2 then 25).
 * Min is enforced by form validation on submit.
 */
export function parseAgeWhileTyping(raw: string, max: number): number {
  const digits = raw.replace(/\D/g, '');
  if (digits === '') return 0;
  let n = parseInt(digits, 10);
  if (!Number.isFinite(n)) return 0;
  if (n < 0) return 0;
  if (n > max) n = max;
  return n;
}

/** Show blank when value is 0 (unset) so the field can start empty. */
export function formatOptionalNumberDisplay(value: number): string {
  if (value === 0) return '';
  return String(value);
}
