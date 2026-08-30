const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8081/api/v1";

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
  meta?: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
}

let isRefreshing = false;

export async function fetchApi<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<ApiResponse<T>> {
  let token = typeof window !== "undefined" ? localStorage.getItem("kresconet_token") : null;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  try {
    let res = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      credentials: "include",
      headers,
    });

    if (
      res.status === 401 &&
      !endpoint.includes("/auth/otp") &&
      !endpoint.includes("/auth/refresh") &&
      !isRefreshing &&
      typeof window !== "undefined"
    ) {
      isRefreshing = true;
      try {
        const refreshRes = await fetch(`${API_BASE}/auth/refresh`, {
          method: "POST",
          credentials: "include",
          headers: { "Content-Type": "application/json" },
        });
        const refreshData = await refreshRes.json();
        isRefreshing = false;

        const newToken =
          refreshData.access_token ||
          refreshData.token ||
          (refreshData.data && (refreshData.data.access_token || refreshData.data.token));

        if (refreshRes.ok && newToken) {
          localStorage.setItem("kresconet_token", newToken);
          headers["Authorization"] = `Bearer ${newToken}`;
          res = await fetch(`${API_BASE}${endpoint}`, {
            ...options,
            credentials: "include",
            headers,
          });
        } else {
          localStorage.removeItem("kresconet_token");
          localStorage.removeItem("kresconet_refresh_token");
        }
      } catch {
        isRefreshing = false;
        localStorage.removeItem("kresconet_token");
        localStorage.removeItem("kresconet_refresh_token");
      }
    }

    const data = await res.json();
    return data as ApiResponse<T>;
  } catch (err: any) {
    return {
      success: false,
      error: {
        code: "NETWORK_ERROR",
        message: err.message || "Failed to connect to backend server",
      },
    };
  }
}
