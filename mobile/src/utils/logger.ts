const LOG_PREFIX = '[FPReg]';

export const logger = {
  info(tag: string, message: string, data?: Record<string, unknown>) {
    console.log(`${LOG_PREFIX} [${tag}] ${message}`, data ?? '');
  },

  warn(tag: string, message: string, data?: Record<string, unknown>) {
    console.warn(`${LOG_PREFIX} [${tag}] ${message}`, data ?? '');
  },

  error(tag: string, message: string, error?: unknown) {
    const safe = error instanceof Error ? { message: error.message } : error;
    console.error(`${LOG_PREFIX} [${tag}] ${message}`, safe ?? '');
  },

  auth(action: string, details?: string) {
    console.log(`${LOG_PREFIX} [Auth] ${action}${details ? ': ' + details : ''}`);
  },

  api(method: string, url: string, status?: number) {
    console.log(`${LOG_PREFIX} [API] ${method} ${url}${status ? ' → ' + status : ''}`);
  },
};
