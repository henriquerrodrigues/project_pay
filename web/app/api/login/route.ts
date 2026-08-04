import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { AUTH_COOKIE, apiUrl, apiErrorMessage } from "@/lib/api";

export async function POST(req: Request) {
  const body = await req.json();

  const res = await fetch(apiUrl("/auth/login"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    cache: "no-store",
  });

  if (!res.ok) {
    const message = await apiErrorMessage(res, "credenciais inválidas");
    return NextResponse.json({ message }, { status: res.status });
  }

  const { token } = await res.json();

  (await cookies()).set(AUTH_COOKIE, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 60 * 60 * 24, // alinhado às 24h do JWT do Go
  });

  return NextResponse.json({ ok: true });
}
