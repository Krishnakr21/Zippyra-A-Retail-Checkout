export class ApiError extends Error {
  code: string;
  requestId?: string;
  status: number;

  constructor(code: string, message: string, status: number, requestId?: string) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;
    this.requestId = requestId;
  }
}

export async function fetchBFF<T>(url: string, options: RequestInit = {}): Promise<T> {
  const headers = {
    'Content-Type': 'application/json',
    ...(options.headers || {}),
  };

  const response = await fetch(url, {
    ...options,
    headers,
  });

  if (!response.ok) {
    let errorData: any = {};
    try {
      errorData = await response.json();
    } catch (_) {}

    const errObj = errorData.error || errorData;
    let code = errObj.code || `HTTP_${response.status}`;
    let message = errObj.message || response.statusText || 'An unexpected error occurred';
    const requestId = errObj.request_id;

    if (response.status === 429) {
      code = 'TOO_MANY_REQUESTS';
      if (!errObj.message || errObj.message === 'Too Many Requests') {
        message = 'Too many requests. Please try again shortly.';
      }
    }

    throw new ApiError(code, message, response.status, requestId);
  }

  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
}

export const api = {
  get: <T>(url: string, options: RequestInit = {}) => fetchBFF<T>(url, { method: 'GET', ...options }),
  post: <T>(url: string, body?: any, options: RequestInit = {}) =>
    fetchBFF<T>(url, { ...options, method: 'POST', body: body ? JSON.stringify(body) : undefined }),
  put: <T>(url: string, body?: any, options: RequestInit = {}) =>
    fetchBFF<T>(url, { ...options, method: 'PUT', body: body ? JSON.stringify(body) : undefined }),
  delete: <T>(url: string, body?: any, options: RequestInit = {}) =>
    fetchBFF<T>(url, { ...options, method: 'DELETE', body: body ? JSON.stringify(body) : undefined }),
};
