const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080/api/v1';

export function getAuthToken(): string | null {
  return localStorage.getItem('kresconet_admin_token');
}

export function setAuthToken(token: string) {
  localStorage.setItem('kresconet_admin_token', token);
}

export function clearAuthToken() {
  localStorage.removeItem('kresconet_admin_token');
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getAuthToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

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
    request<{ access_token: string; token?: string; refresh_token?: string; user: { id: string; name: string; role: string } }>('/auth/employee/login', {
      method: 'POST',
      body: JSON.stringify({ email: 'kresconet@gmail.com', password }),
    }),

  loginWithCredentials: (email: string, password: string) =>
    request<{ access_token: string; token?: string; refresh_token?: string; user: { id: string; name: string; role: string } }>('/auth/employee/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
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

  approveApplication: (id: string, reason?: string) =>
    request<any>(`/admin/applications/${id}/approve`, {
      method: 'POST',
      body: JSON.stringify({ reason }),
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

  // Verification & Credit Auto-Trigger
  triggerVerifications: (appId: string, distId: string) =>
    request<any>(`/verifications/trigger-all?application_id=${appId}&distributor_id=${distId}`, {
      method: 'POST',
    }),

  evaluateCredit: (appId: string) =>
    request<any>(`/credit/evaluate?application_id=${appId}`, {
      method: 'POST',
    }),

  // Distributors
  listDistributors: (limit = 50, offset = 0) =>
    request<any>(`/distributors?limit=${limit}&offset=${offset}`).then((res) => (Array.isArray(res) ? { distributors: res, total: res.length } : res)).catch(() => ({ distributors: [], total: 0 })),

  getDistributor: (id: string) =>
    request<any>(`/distributors/${id}`),

  // Policy
  getPolicy: () =>
    request<any>('/admin/policy'),

  reloadPolicy: () =>
    request<any>('/admin/policy/reload', { method: 'POST' }),

  // Orders
  listOrders: (limit = 50, offset = 0) =>
    request<any>(`/orders?limit=${limit}&offset=${offset}`).catch(() => ({ orders: [], total: 0 })),

  approveOrder: (id: string) =>
    request<any>(`/orders/${id}/approve`, { method: 'POST' }),

  dispatchOrder: (id: string) =>
    request<any>(`/orders/${id}/dispatch`, { method: 'POST' }),
};
