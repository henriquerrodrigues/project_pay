import { NextResponse } from "next/server";
import { apiUrl, apiErrorMessage } from "@/lib/api";

export async function POST(req: Request) {
  const body = await req.json();

  const res = await fetch(apiUrl("/users"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    cache: "no-store",
  });

  if (!res.ok) {
    const message = await apiErrorMessage(res, "erro ao criar usuário");
    return NextResponse.json({ message }, { status: res.status });
  }

  return NextResponse.json(await res.json(), { status: 201 });
}
