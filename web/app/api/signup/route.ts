import { NextResponse } from "next/server";
import { apiUrl, apiErrorMessage } from "@/lib/api";

export async function POST(req: Request) {
  const { password, ...rest } = await req.json();

  // o Go espera o campo "password_hash" no body de POST /users
  // (o nome é do contrato existente, não que o cliente mande um hash pronto)
  const res = await fetch(apiUrl("/users"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ...rest, password_hash: password }),
    cache: "no-store",
  });

  if (!res.ok) {
    const message = await apiErrorMessage(res, "erro ao criar usuário");
    return NextResponse.json({ message }, { status: res.status });
  }

  return NextResponse.json(await res.json(), { status: 201 });
}
