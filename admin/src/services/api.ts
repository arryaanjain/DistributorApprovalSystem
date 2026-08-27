const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080/api/v1';

export function getAuthToken(): string | null {
  return localStorage.getItem('kresconet_admin_token');
}

export function getRefreshToken(): string | null {
  return localStorage.getItem('kresconet_admin_refresh_token');
}

export function setAuthToken(token: string, refreshToken?: string) {
  localStorage.setItem('kresconet_admin_token', token);
  if (refreshToken) {
    localStorage.setItem('kresconet_admin_refresh_token', refreshToken);
  }
}

export function clearAuthToken() {
  localStorage.removeItem('kresconet_admin_token');
  localStorage.removeItem('kresconet_admin_refresh_token');
}

let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string) => void;
  reject: (err: any) => void;
}> = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else if (token) {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getAuthToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  let res = await fetch(`${API_BASE}${path}`, {
    ...options,
    credentials: 'include',
    headers,
  });

  if (res.status === 401 && !path.includes('/login') && !path.includes('/refresh')) {
    if (isRefreshing) {
      return new Promise<string>((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      }).then((newToken) => {
        headers['Authorization'] = `Bearer ${newToken}`;
        return fetch(`${API_BASE}${path}`, {
          ...options,
          credentials: 'include',
          headers,
        }).then(async (retryRes) => {
          const body = await retryRes.json().catch(() => ({}));
          if (!retryRes.ok) {
            const errorMsg = body?.error?.message || body?.message || retryRes.statusText || 'API Request Failed';
            throw new Error(errorMsg);
          }
          return body.data !== undefined ? body.data : body;
        });
      });
    }

    isRefreshing = true;
    try {
      const storedRefreshToken = getRefreshToken();
      const refreshRes = await fetch(`${API_BASE}/auth/employee/refresh`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: storedRefreshToken || '' }),
      });
      const refreshData = await refreshRes.json();
      isRefreshing = false;
      if (refreshRes.ok && (refreshData.access_token || refreshData.token)) {
        const newAccess = refreshData.access_token || refreshData.token;
        const newRefresh = refreshData.refresh_token;
        setAuthToken(newAccess, newRefresh);
        processQueue(null, newAccess);
        headers['Authorization'] = `Bearer ${newAccess}`;
        res = await fetch(`${API_BASE}${path}`, {
          ...options,
          credentials: 'include',
          headers,
        });
      } else {
        const err = new Error('Session expired. Please log in again.');
        processQueue(err, null);
        clearAuthToken();
        localStorage.removeItem('kresconet_admin_user');
        window.location.href = '/';
        throw err;
      }
    } catch (err) {
      isRefreshing = false;
      processQueue(err, null);
      clearAuthToken();
      localStorage.removeItem('kresconet_admin_user');
      window.location.href = '/';
      throw err;
    }
  }

  const body = await res.json().catch(() => ({}));

  if (!res.ok) {
    const errorMsg = body?.error?.message || body?.message || res.statusText || 'API Request Failed';
    throw new Error(errorMsg);
  }

  return body.data !== undefined ? body.data : body;
}

export const api = {
  // Auth
  login: (password: string) =>
    request<{ access_token: string; refresh_token?: string; token?: string; user: { id: string; name: string; role: string } }>('/auth/employee/login', {
      method: 'POST',
      body: JSON.stringify({ email: 'admin@kresconet.com', password }),
    }),

  loginWithCredentials: (email: string, password: string) =>
    request<{ access_token: string; refresh_token?: string; token?: string; user: { id: string; name: string; role: string } }>('/auth/employee/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  logout: () =>
    request('/auth/logout', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: getRefreshToken() || '' }),
    }),

  // Products & Sample Catalogue
  listProductsAdmin: () =>
    request<any[]>('/catalogue/admin'),

  createProduct: (p: any) =>
    request<any>('/catalogue', {
      method: 'POST',
      body: JSON.stringify(p),
    }),

  updateProduct: (id: string, p: any) =>
    request<any>(`/catalogue/${id}`, {
      method: 'PUT',
      body: JSON.stringify(p),
    }),

  // Applications
  listApplications: (status: string = 'all', limit = 50, offset = 0) =>
    request<{ applications: any[]; total: number }>(`/admin/applications?status=${status}&limit=${limit}&offset=${offset}`).catch(() => ({ applications: [], total: 0 })),

  getApplication: (id: string) =>
    request<any>(`/admin/applications/${id}`),

  approveApplication: (id: string, reason?: string, approvedLimitPaise?: number, approvedPeriodDays?: number) =>
    request<any>(`/admin/applications/${id}/approve`, {
      method: 'POST',
      body: JSON.stringify({
        reason,
        approved_limit_paise: approvedLimitPaise,
        approved_period_days: approvedPeriodDays,
      }),
    }),

  rejectApplication: (id: string, reason?: string) =>
    request<any>(`/admin/applications/${id}/reject`, {
      method: 'POST',
      body: JSON.stringify({ reason }),
    }),

  holdApplication: (id: string, reason?: string) =>
    request<any>(`/admin/applications/${id}/hold`, {
      method: 'POST',
      body: JSON.stringify({ reason }),
    }),

  // Dashboard Stats
  getDashboardStats: () =>
    request<{
      total_applications: number;
      pending_verifications: number;
      total_distributors: number;
      sanctioned_credit_paise: number;
      utilized_credit_paise: number;
      available_credit_paise: number;
    }>('/admin/dashboard-stats'),

  // Verification & Credit Auto-Trigger
  triggerVerifications: (appId: string, distId: string) =>
    request<any>(`/verification/${appId}/trigger?distributor_id=${distId}`, {
      method: 'POST',
    }),

  evaluateCredit: (appId: string) =>
    request<any>(`/credit/${appId}/evaluate`, {
      method: 'POST',
    }),

  // Distributors
  listDistributors: (limit = 50, offset = 0) =>
    request<any>(`/distributors?limit=${limit}&offset=${offset}`).then((res) => (Array.isArray(res) ? { distributors: res, total: res.length } : res)).catch(() => ({ distributors: [], total: 0 })),

  getDistributor: (id: string) =>
    request<any>(`/distributors/${id}`),

  getDistributorCreditTrail: (id: string) =>
    request<{ distributor_id: string; trail: any[] }>(`/admin/distributors/${id}/credit-trail`),

  // Policy
  getPolicy: () =>
    request<any>('/admin/policy'),

  reloadPolicy: () =>
    request<any>('/admin/policy/reload', { method: 'POST' }),

  // Orders
  listOrders: (limit = 50, offset = 0) =>
    request<any>(`/orders/all?limit=${limit}&offset=${offset}`).catch(() => ({ orders: [], total: 0 })),

  listCatalogOrders: (limit = 50, offset = 0) =>
    request<any>(`/orders/all?limit=${limit}&offset=${offset}`).catch(() => ({ orders: [], total: 0 })),

  listSampleOrders: (limit = 50, offset = 0) =>
    request<any>(`/orders/samples?limit=${limit}&offset=${offset}`).catch(() => ({ sample_orders: [], total: 0 })),

  approveOrder: (id: string) =>
    request<any>(`/orders/${id}/approve`, { method: 'POST' }),

  dispatchOrder: (id: string) =>
    request<any>(`/orders/${id}/dispatch`, { method: 'POST' }),

  dispatchSampleOrder: (sampleOrderId: string) =>
    request<any>('/shipping/sample-dispatch', {
      method: 'POST',
      body: JSON.stringify({ sample_order_id: sampleOrderId }),
    }),

  // Shiprocket Logistics APIs
  createShipment: (orderId: string, payload: { weight: number; length: number; breadth: number; height: number; payment_method: string; pickup_location: string }) =>
    request<any>(`/shipping/create/${orderId}`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  getWalletBalance: () =>
    request<any>('/shipping/wallet-balance'),

  getAvailableCouriers: (orderId: string) =>
    request<any>(`/shipping/couriers/${orderId}`),

  assignCourier: (orderId: string, courierId: number | string, courierRate: number) =>
    request<any>(`/shipping/assign-courier/${orderId}`, {
      method: 'POST',
      body: JSON.stringify({ courier_id: courierId, courier_rate: courierRate }),
    }),

  requestPickup: (orderId: string) =>
    request<any>(`/shipping/pickup/${orderId}`, {
      method: 'POST',
    }),

  generateLabel: (orderId: string) =>
    request<any>(`/shipping/label/${orderId}`),

  generateManifest: (orderId: string) =>
    request<any>(`/shipping/manifest/${orderId}`, {
      method: 'POST',
    }),

  trackShipment: (orderId: string) =>
    request<any>(`/shipping/track/${orderId}`),
};
