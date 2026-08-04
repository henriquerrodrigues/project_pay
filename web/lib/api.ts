// Helpers server-side. O JWT vive num cookie httpOnly: o browser nunca o vê;
// todo acesso à API Go passa por aqui (Server Components, Server Actions,
// route handlers).
import { cookies } from "next/headers";

export const AUTH_COOKIE = "pp_token";

const API_URL = process.env.API_URL ?? "http://localhost:8080";

export function apiUrl(path: string): string {
  return `${API_URL}${path}`;
}

// Injeta Authorization: Bearer <jwt> lido do cookie e desliga cache.
export async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const token = (await cookies()).get(AUTH_COOKIE)?.value;

  const headers = new Headers(init.headers);
  headers.set("Content-Type", "application/json");
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  return fetch(apiUrl(path), { ...init, headers, cache: "no-store" });
}

// Extrai a mensagem do contrato de erro; fallback genérico se o body não for JSON.
export async function apiErrorMessage(res: Response, fallback: string): Promise<string> {
  try {
    const body = await res.json();
    return typeof body?.message === "string" ? body.message : fallback;
  } catch {
    return fallback;
  }
}
