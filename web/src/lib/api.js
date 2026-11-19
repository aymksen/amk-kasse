export class ApiError extends Error {
  constructor(code, status) {
    super(code);
    this.code = code;
    this.status = status;
  }
}

let onSessionLost = () => {};

/** Runs when the server says the admin session is gone. */
export function whenSessionLost(fn) {
  onSessionLost = fn;
}

async function request(method, url, body) {
  const isJson = body !== undefined && !(body instanceof Blob);
  let res;
  try {
    res = await fetch(url, {
      method,
      headers: isJson ? { 'Content-Type': 'application/json' } : undefined,
      body: isJson ? JSON.stringify(body) : body,
    });
  } catch {
    throw new ApiError('offline', 0);
  }
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    if (data?.error === 'unauthorized') onSessionLost();
    throw new ApiError(data?.error ?? 'internal', res.status);
  }
  return data;
}

export const get = (url) => request('GET', url);
export const post = (url, body = {}) => request('POST', url, body);
export const upload = (url, blob) => request('POST', url, blob);
export const remove = (url) => request('DELETE', url);
