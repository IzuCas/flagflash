// FlagFlash API - Auth Service

// Get API URL from environment or use proxy path for dev
const API_URL = '/api/v1/flagflash';
const DEV = import.meta.env.DEV;

function getAuthHeader(): Record<string, string> {
  const token = localStorage.getItem('auth_token');
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function request<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_URL}${endpoint}`;
  if (DEV) console.debug(`[API] ${options.method || 'GET'} ${endpoint}`);

  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
        ...options.headers,
      },
    });

    if (DEV) console.debug(`[API] Response: ${response.status} ${response.statusText}`);

    if (response.status === 401) {
      localStorage.removeItem('auth_token');
      localStorage.removeItem('auth_username');
      window.location.href = '/login';
      throw new Error('Unauthorized');
    }

    if (!response.ok) {
      const error = await response.json().catch(() => ({ detail: 'Unknown error' }));
      if (DEV) console.error('[API] Error response:', error);
      throw new Error(error.detail || error.message || 'Request failed');
    }

    if (response.status === 204) {
      return {} as T;
    }

    const data = await response.json();
    if (DEV) console.debug(`[API] Data received:`, Array.isArray(data) ? `Array(${data.length})` : typeof data);
    return data;
  } catch (error) {
    if (DEV) console.error(`[API] Fetch error for ${endpoint}:`, error);
    throw error;
  }
}

// Auth API (no token required)
export const authApi = {
  login: (username: string, password: string) =>
    fetch(`${API_URL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    }).then(async (res) => {
      if (!res.ok) {
        const err = await res.json().catch(() => ({ detail: 'Login failed' }));
        throw new Error(err.detail || err.message || 'Invalid username or password');
      }
      return res.json() as Promise<{ token: string; username: string; require_password_change: boolean }>;
    }),

  changePassword: (currentPassword: string, newPassword: string, newUsername?: string) =>
    request<{ message: string }>('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify({
        current_password: currentPassword,
        new_password: newPassword,
        ...(newUsername ? { new_username: newUsername } : {}),
      }),
    }),

  verify: (username: string, password: string) =>
    request<{ valid: boolean }>('/auth/verify', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),
};
