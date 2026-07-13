# Semana 4 — Frontend mínimo (Next.js) — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Frontend Next.js em `web/` com login via cookie httpOnly (proxy pro Go), signup, dashboard com contas/saldo/transferência e extrato — fluxo completo funcionando no navegador.

**Architecture:** Next.js 15 App Router. O browser nunca vê o JWT: route handlers (`/api/login`, `/api/logout`, `/api/signup`) fazem proxy pra API Go e guardam o token em cookie httpOnly. Server Components buscam dados com o helper `apiFetch` (injeta `Authorization: Bearer` a partir do cookie). Mutações (transferência, nova conta) via Server Actions + `revalidatePath`. `middleware.ts` redireciona rota protegida sem cookie para `/login`.

**Tech Stack:** Next.js 15 (App Router) + TypeScript + Tailwind v4 + shadcn/ui. API Go da Semana 3 rodando em `http://localhost:8080`.

**Pré-requisito:** plano `2026-07-13-semana-3-backend.md` executado (contrato de erro `{"code","message","trace_id"}`, `GET /accounts`, `POST /accounts` sem `user_id` no body, `POST /users` com campo `password`).

**Fora de escopo:** testes E2E, dark mode, responsividade caprichada, deploy. Verificação = `npm run build` + fluxo manual no navegador.

**Adições ao spec (lacunas detectadas):**
1. Dialog "Nova conta" no dashboard — o fluxo de aceite exige "ver contas" mas nenhuma tela do spec cria conta.
2. Seed de dev (`scripts/dev_seed.sql`) — conta nova nasce com `balance = 0` e não existe endpoint de depósito; sem crédito manual não há transferência possível no fluxo de aceite.

---

## Estrutura de arquivos (visão geral)

| Arquivo | Responsabilidade |
|---|---|
| `scripts/dev_seed.sql` | Participante fixo de dev + exemplo de crédito de saldo |
| `web/.env.local` | `API_URL` da API Go (server-side only) |
| `web/lib/api.ts` | `apiFetch` + nome do cookie |
| `web/lib/types.ts` | Tipos espelhando o JSON da API Go |
| `web/lib/money.ts` | Centavos ↔ BRL |
| `web/app/api/{login,logout,signup}/route.ts` | Proxy de auth + cookie httpOnly |
| `web/middleware.ts` | Proteção de rota |
| `web/app/page.tsx` | Redirect `/` → `/dashboard` |
| `web/app/login/page.tsx`, `web/app/signup/page.tsx` | Forms públicos (client) |
| `web/app/dashboard/page.tsx` | Server Component: lista contas |
| `web/app/dashboard/actions.ts` | Server Actions: transferir, criar conta |
| `web/app/dashboard/transfer-dialog.tsx`, `new-account-dialog.tsx`, `logout-button.tsx` | Client components |
| `web/app/accounts/[id]/statement/page.tsx` + `statement-table.tsx` | Extrato + "carregar mais" (client-side) |

Comandos rodam da raiz do repo, salvo indicação de `web/`.

---

### Task 1: Seed de dev

**Files:**
- Create: `scripts/dev_seed.sql`

- [ ] **Step 1: Criar o seed**

```sql
-- scripts/dev_seed.sql
-- Seed de desenvolvimento. Rode após as migrations:
--   docker exec -i project_pay_db psql -U pay_user -d project_pay < scripts/dev_seed.sql
--
-- Participante fixo usado como default no form "Nova conta" do frontend.
INSERT INTO participants (id, ispb, name)
VALUES ('00000000-0000-0000-0000-000000000001', '12345678', 'Banco Dev')
ON CONFLICT (id) DO NOTHING;

-- Contas novas nascem com balance = 0 e não existe endpoint de depósito.
-- Para testar transferência, credite saldo manualmente (valor em centavos):
--   UPDATE accounts SET balance = balance + 100000 WHERE id = '<ACCOUNT_ID>';
```

- [ ] **Step 2: Aplicar e verificar**

Run: `docker-compose up -d postgres` e depois `docker exec -i project_pay_db psql -U pay_user -d project_pay < scripts/dev_seed.sql`
Expected: `INSERT 0 1` (ou `INSERT 0 0` se já existir).

- [ ] **Step 3: Commit**

```bash
git add scripts/dev_seed.sql
git commit -m "chore: seed de dev com participante fixo"
```

---

### Task 2: Scaffold do Next.js + shadcn/ui

**Files:**
- Create: `web/` (via create-next-app)
- Create: `web/.env.local`

- [ ] **Step 1: Criar o app**

Na raiz do repo:

```bash
npx create-next-app@latest web --typescript --tailwind --eslint --app --no-src-dir --import-alias "@/*" --use-npm
```

(Se perguntar sobre Turbopack, aceite o default.)

- [ ] **Step 2: Inicializar shadcn/ui e adicionar componentes**

```bash
cd web
npx shadcn@latest init -d
npx shadcn@latest add button card dialog input label select table
```

- [ ] **Step 3: Variável de ambiente**

```bash
# web/.env.local
API_URL=http://localhost:8080
```

Sem prefixo `NEXT_PUBLIC_` — o valor só é usado server-side; o browser nunca fala direto com o Go.

Confirmar que `.env*` já está no `web/.gitignore` gerado pelo create-next-app (está, por padrão).

- [ ] **Step 4: Build de verificação**

Run: `npm run build` (dentro de `web/`)
Expected: build verde.

- [ ] **Step 5: Commit**

```bash
git add web/
git commit -m "feat: scaffold do frontend Next.js 15 com Tailwind e shadcn/ui"
```

---

### Task 3: Helpers — tipos, dinheiro e `apiFetch`

**Files:**
- Create: `web/lib/types.ts`
- Create: `web/lib/money.ts`
- Create: `web/lib/api.ts`

- [ ] **Step 1: Tipos do contrato da API**

Os nomes espelham exatamente as tags JSON dos models Go (accounts/transfers em camelCase, ledger em snake_case — o contrato é assim hoje):

```ts
// web/lib/types.ts
export type Account = {
  id: string;
  userID: string;
  participantID: string;
  accountNumber: string;
  branch: string;
  accountType: "cacc" | "svgs";
  balance: number; // centavos
  createdAt: string;
};

export type Transfer = {
  id: string;
  endtoendID: string;
  senderAccountID: string;
  receiverAccountID: string;
  amount: number; // centavos
  status: string;
  description: string;
  createdAt: string;
  completedAt: string | null;
};

export type LedgerEntry = {
  id: string;
  transfer_id: string;
  account_id: string;
  amount: number; // centavos
  entry_type: "debit" | "credit";
  balance_after: number; // centavos
  created_at: string;
};

// contrato de erro da API (apperrors, Semana 3)
export type ApiError = {
  code: string;
  message: string;
  trace_id: string;
};
```

- [ ] **Step 2: Formatação de dinheiro**

```ts
// web/lib/money.ts
const brl = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
});

// A API trabalha em centavos (int64).
export function formatBRL(cents: number): string {
  return brl.format(cents / 100);
}

// "12,50" ou "12.50" -> 1250. Retorna NaN se inválido.
export function parseBRLToCents(input: string): number {
  const value = Number(input.replace(",", "."));
  return Number.isFinite(value) ? Math.round(value * 100) : NaN;
}
```

- [ ] **Step 3: `apiFetch`**

```ts
// web/lib/api.ts
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
```

- [ ] **Step 4: Build**

Run: `npm run build` (em `web/`)
Expected: verde.

- [ ] **Step 5: Commit**

```bash
git add web/lib/
git commit -m "feat: helpers de API, tipos do contrato e formatacao BRL"
```

---

### Task 4: Route handlers de auth + proteção de rota

**Files:**
- Create: `web/app/api/login/route.ts`
- Create: `web/app/api/logout/route.ts`
- Create: `web/app/api/signup/route.ts`
- Create: `web/middleware.ts`
- Modify: `web/app/page.tsx`

- [ ] **Step 1: Login (grava o cookie httpOnly)**

```ts
// web/app/api/login/route.ts
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
```

- [ ] **Step 2: Logout (apaga o cookie)**

```ts
// web/app/api/logout/route.ts
import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { AUTH_COOKIE } from "@/lib/api";

export async function POST() {
  (await cookies()).delete(AUTH_COOKIE);
  return NextResponse.json({ ok: true });
}
```

- [ ] **Step 3: Signup (proxy do `POST /users`, rota pública)**

```ts
// web/app/api/signup/route.ts
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
```

- [ ] **Step 4: Middleware de proteção**

```ts
// web/middleware.ts
import { NextResponse, type NextRequest } from "next/server";

const PUBLIC_PATHS = ["/login", "/signup", "/api/login", "/api/signup"];

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;

  const isPublic = PUBLIC_PATHS.some(
    (p) => pathname === p || pathname.startsWith(`${p}/`)
  );
  if (isPublic) {
    return NextResponse.next();
  }

  const hasToken = Boolean(req.cookies.get("pp_token")?.value);
  if (!hasToken) {
    return NextResponse.redirect(new URL("/login", req.url));
  }

  return NextResponse.next();
}

// tudo menos assets do Next
export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
```

(O nome do cookie fica literal aqui porque o middleware roda no edge runtime e não deve importar `lib/api.ts`, que puxa `next/headers`.)

- [ ] **Step 5: Redirect da raiz**

Substituir `web/app/page.tsx` por:

```tsx
// web/app/page.tsx
import { redirect } from "next/navigation";

export default function Home() {
  redirect("/dashboard");
}
```

- [ ] **Step 6: Build + verificação manual**

Run: `npm run build` (em `web/`)
Expected: verde.

Manual: `npm run dev`, abrir `http://localhost:3000/` → deve redirecionar para `/login` (404 da página é esperado — ela nasce na Task 5; o redirect é o que importa).

- [ ] **Step 7: Commit**

```bash
git add web/app/api/ web/middleware.ts web/app/page.tsx
git commit -m "feat: auth por cookie httpOnly com proxy e protecao de rota"
```

---

### Task 5: Tela `/login`

**Files:**
- Create: `web/app/login/page.tsx`

- [ ] **Step 1: Implementar a página (client component)**

```tsx
// web/app/login/page.tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setPending(true);
    setError(null);

    const res = await fetch("/api/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });

    setPending(false);
    if (!res.ok) {
      const body = await res.json().catch(() => null);
      setError(body?.message ?? "credenciais inválidas");
      return;
    }
    router.push("/dashboard");
  }

  return (
    <main className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Entrar</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Senha</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            {error && <p className="text-sm text-red-600">{error}</p>}
            <Button type="submit" className="w-full" disabled={pending}>
              {pending ? "Entrando..." : "Entrar"}
            </Button>
            <p className="text-sm text-muted-foreground">
              Não tem conta? <Link href="/signup" className="underline">Cadastre-se</Link>
            </p>
          </form>
        </CardContent>
      </Card>
    </main>
  );
}
```

- [ ] **Step 2: Build + verificação manual**

Run: `npm run build` (em `web/`)
Expected: verde.

Manual (API Go + Postgres de pé): criar um user via curl (`POST /users`), abrir `/login`, logar → redireciona pra `/dashboard` (404 por enquanto — nasce na Task 6). Senha errada → mensagem de erro no form. Conferir no DevTools que o cookie `pp_token` é `HttpOnly`.

- [ ] **Step 3: Commit**

```bash
git add web/app/login/
git commit -m "feat: tela de login"
```

---

### Task 6: Tela `/signup`

**Files:**
- Create: `web/app/signup/page.tsx`

- [ ] **Step 1: Implementar a página**

```tsx
// web/app/signup/page.tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function SignupPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [taxID, setTaxID] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setPending(true);
    setError(null);

    const res = await fetch("/api/signup", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      // tax_id/password: nomes do contrato do POST /users do Go
      body: JSON.stringify({ name, email, tax_id: taxID, password }),
    });

    setPending(false);
    if (!res.ok) {
      const body = await res.json().catch(() => null);
      setError(body?.message ?? "erro ao criar usuário");
      return;
    }
    router.push("/login");
  }

  return (
    <main className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Criar conta</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Nome</Label>
              <Input id="name" value={name} onChange={(e) => setName(e.target.value)} required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input id="email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="taxID">CPF</Label>
              <Input id="taxID" value={taxID} onChange={(e) => setTaxID(e.target.value)} required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Senha</Label>
              <Input id="password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
            </div>
            {error && <p className="text-sm text-red-600">{error}</p>}
            <Button type="submit" className="w-full" disabled={pending}>
              {pending ? "Criando..." : "Cadastrar"}
            </Button>
            <p className="text-sm text-muted-foreground">
              Já tem conta? <Link href="/login" className="underline">Entrar</Link>
            </p>
          </form>
        </CardContent>
      </Card>
    </main>
  );
}
```

- [ ] **Step 2: Build + verificação manual**

Run: `npm run build` (em `web/`)
Expected: verde.

Manual: `/signup` → cadastrar → redireciona pra `/login` → logar com as credenciais novas funciona.

- [ ] **Step 3: Commit**

```bash
git add web/app/signup/
git commit -m "feat: tela de cadastro"
```

---

### Task 7: Dashboard — Server Actions + página com contas

**Files:**
- Create: `web/app/dashboard/actions.ts`
- Create: `web/app/dashboard/logout-button.tsx`
- Create: `web/app/dashboard/page.tsx`
- Create: `web/app/dashboard/transfer-dialog.tsx` (stub; código real na Task 8)
- Create: `web/app/dashboard/new-account-dialog.tsx` (stub; código real na Task 8)

- [ ] **Step 1: Server Actions**

```ts
// web/app/dashboard/actions.ts
"use server";

import { revalidatePath } from "next/cache";
import { apiFetch, apiErrorMessage } from "@/lib/api";

export type ActionResult = { ok: boolean; message?: string };

export async function createTransfer(input: {
  senderAccountID: string;
  receiverAccountID: string;
  amount: number; // centavos
  description: string;
}): Promise<ActionResult> {
  const res = await apiFetch("/transfers", {
    method: "POST",
    body: JSON.stringify(input),
  });

  if (!res.ok) {
    return { ok: false, message: await apiErrorMessage(res, "erro ao transferir") };
  }

  revalidatePath("/dashboard");
  return { ok: true };
}

export async function createAccount(input: {
  participantID: string;
  accountNumber: string;
  branch: string;
  accountType: string;
}): Promise<ActionResult> {
  const res = await apiFetch("/accounts", {
    method: "POST",
    body: JSON.stringify(input),
  });

  if (!res.ok) {
    return { ok: false, message: await apiErrorMessage(res, "erro ao criar conta") };
  }

  revalidatePath("/dashboard");
  return { ok: true };
}
```

- [ ] **Step 2: Botão de logout**

```tsx
// web/app/dashboard/logout-button.tsx
"use client";

import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";

export function LogoutButton() {
  const router = useRouter();

  async function handleLogout() {
    await fetch("/api/logout", { method: "POST" });
    router.push("/login");
  }

  return (
    <Button variant="outline" onClick={handleLogout}>
      Sair
    </Button>
  );
}
```

- [ ] **Step 3: Página do dashboard (Server Component)**

Os dialogs importados aqui nascem na Task 8 — para a página compilar já nesta task, crie os dois arquivos como stub e substitua na Task 8:

```tsx
// web/app/dashboard/transfer-dialog.tsx  (stub — substituído na Task 8)
"use client";
import type { Account } from "@/lib/types";

export function TransferDialog({ accounts }: { accounts: Account[] }) {
  return null;
}
```

```tsx
// web/app/dashboard/new-account-dialog.tsx  (stub — substituído na Task 8)
"use client";

export function NewAccountDialog() {
  return null;
}
```

```tsx
// web/app/dashboard/page.tsx
import Link from "next/link";
import { redirect } from "next/navigation";
import { apiFetch } from "@/lib/api";
import { formatBRL } from "@/lib/money";
import type { Account } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { TransferDialog } from "./transfer-dialog";
import { NewAccountDialog } from "./new-account-dialog";
import { LogoutButton } from "./logout-button";

export default async function DashboardPage() {
  const res = await apiFetch("/accounts");
  if (res.status === 401) {
    redirect("/login"); // cookie presente mas token expirado/inválido
  }
  if (!res.ok) {
    throw new Error("erro ao carregar contas");
  }
  const accounts: Account[] = await res.json();

  return (
    <main className="mx-auto max-w-3xl space-y-6 p-6">
      <header className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Minhas contas</h1>
        <div className="flex gap-2">
          <NewAccountDialog />
          <TransferDialog accounts={accounts} />
          <LogoutButton />
        </div>
      </header>

      {accounts.length === 0 && (
        <p className="text-muted-foreground">
          Você ainda não tem contas. Crie uma em &quot;Nova conta&quot;.
        </p>
      )}

      <div className="grid gap-4 sm:grid-cols-2">
        {accounts.map((account) => (
          <Card key={account.id}>
            <CardHeader>
              <CardTitle className="text-base">
                Conta {account.accountNumber} · Ag. {account.branch}
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <p className="text-2xl font-semibold">{formatBRL(account.balance)}</p>
              <Link
                href={`/accounts/${account.id}/statement`}
                className="text-sm underline"
              >
                Ver extrato
              </Link>
            </CardContent>
          </Card>
        ))}
      </div>
    </main>
  );
}
```

- [ ] **Step 4: Build + verificação manual**

Run: `npm run build` (em `web/`)
Expected: verde.

Manual: logar → dashboard mostra "Você ainda não tem contas" (user novo) e o botão Sair volta pro login (cookie apagado; `/dashboard` redireciona pra `/login`).

- [ ] **Step 5: Commit**

```bash
git add web/app/dashboard/
git commit -m "feat: dashboard com listagem de contas e logout"
```

---

### Task 8: Dialogs de nova conta e transferência

**Files:**
- Modify: `web/app/dashboard/new-account-dialog.tsx` (substitui o stub)
- Modify: `web/app/dashboard/transfer-dialog.tsx` (substitui o stub)

- [ ] **Step 1: Dialog "Nova conta"**

Substituir o stub por:

```tsx
// web/app/dashboard/new-account-dialog.tsx
"use client";

import { useState } from "react";
import { createAccount } from "./actions";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

// Participante do scripts/dev_seed.sql
const DEV_PARTICIPANT_ID = "00000000-0000-0000-0000-000000000001";

export function NewAccountDialog() {
  const [open, setOpen] = useState(false);
  const [participantID, setParticipantID] = useState(DEV_PARTICIPANT_ID);
  const [accountNumber, setAccountNumber] = useState("");
  const [branch, setBranch] = useState("0001");
  const [accountType, setAccountType] = useState("cacc");
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setPending(true);
    setError(null);

    const result = await createAccount({ participantID, accountNumber, branch, accountType });

    setPending(false);
    if (!result.ok) {
      setError(result.message ?? "erro ao criar conta");
      return;
    }
    setOpen(false);
    setAccountNumber("");
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="secondary">Nova conta</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Nova conta</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="participantID">Participante (ID)</Label>
            <Input
              id="participantID"
              value={participantID}
              onChange={(e) => setParticipantID(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="accountNumber">Número da conta</Label>
            <Input
              id="accountNumber"
              value={accountNumber}
              onChange={(e) => setAccountNumber(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="branch">Agência</Label>
            <Input id="branch" value={branch} onChange={(e) => setBranch(e.target.value)} required />
          </div>
          <div className="space-y-2">
            <Label>Tipo</Label>
            <Select value={accountType} onValueChange={setAccountType}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="cacc">Corrente</SelectItem>
                <SelectItem value="svgs">Poupança</SelectItem>
              </SelectContent>
            </Select>
          </div>
          {error && <p className="text-sm text-red-600">{error}</p>}
          <Button type="submit" className="w-full" disabled={pending}>
            {pending ? "Criando..." : "Criar conta"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 2: Dialog de transferência**

Substituir o stub por:

```tsx
// web/app/dashboard/transfer-dialog.tsx
"use client";

import { useState } from "react";
import { createTransfer } from "./actions";
import { parseBRLToCents, formatBRL } from "@/lib/money";
import type { Account } from "@/lib/types";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export function TransferDialog({ accounts }: { accounts: Account[] }) {
  const [open, setOpen] = useState(false);
  const [sender, setSender] = useState("");
  const [receiver, setReceiver] = useState("");
  const [amount, setAmount] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);

    const cents = parseBRLToCents(amount);
    if (!Number.isFinite(cents) || cents <= 0) {
      setError("valor inválido");
      return;
    }

    setPending(true);
    const result = await createTransfer({
      senderAccountID: sender,
      receiverAccountID: receiver,
      amount: cents,
      description,
    });
    setPending(false);

    if (!result.ok) {
      setError(result.message ?? "erro ao transferir");
      return;
    }
    setOpen(false);
    setReceiver("");
    setAmount("");
    setDescription("");
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button disabled={accounts.length === 0}>Transferir</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Nova transferência</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Conta de origem</Label>
            <Select value={sender} onValueChange={setSender} required>
              <SelectTrigger>
                <SelectValue placeholder="Selecione a conta" />
              </SelectTrigger>
              <SelectContent>
                {accounts.map((account) => (
                  <SelectItem key={account.id} value={account.id}>
                    {account.accountNumber} · {formatBRL(account.balance)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="receiver">Conta de destino (ID)</Label>
            <Input
              id="receiver"
              value={receiver}
              onChange={(e) => setReceiver(e.target.value)}
              placeholder="uuid da conta destino"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="amount">Valor (R$)</Label>
            <Input
              id="amount"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="10,00"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="description">Descrição</Label>
            <Input
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          {error && <p className="text-sm text-red-600">{error}</p>}
          <Button type="submit" className="w-full" disabled={pending || !sender}>
            {pending ? "Transferindo..." : "Transferir"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 3: Build + verificação manual**

Run: `npm run build` (em `web/`)
Expected: verde.

Manual: criar conta pelo dialog → aparece no dashboard com R$ 0,00. Creditar saldo via SQL (instrução no `scripts/dev_seed.sql`), refresh → saldo atualizado. Transferir para outra conta → dialog fecha e saldos atualizam (revalidatePath). Transferir mais que o saldo → mensagem "saldo insuficiente" no form (vinda do `message` da API).

- [ ] **Step 4: Commit**

```bash
git add web/app/dashboard/
git commit -m "feat: dialogs de nova conta e transferencia"
```

---

### Task 9: Extrato `/accounts/[id]/statement`

**Files:**
- Create: `web/app/accounts/[id]/statement/page.tsx`
- Create: `web/app/accounts/[id]/statement/statement-table.tsx`

- [ ] **Step 1: Tabela com "carregar mais" (client)**

A API retorna o extrato inteiro (sem paginação no backend — decisão da Semana 3); a paginação "carregar mais" é client-side, de 10 em 10:

```tsx
// web/app/accounts/[id]/statement/statement-table.tsx
"use client";

import { useState } from "react";
import { formatBRL } from "@/lib/money";
import type { LedgerEntry } from "@/lib/types";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const PAGE_SIZE = 10;

export function StatementTable({ entries }: { entries: LedgerEntry[] }) {
  const [visible, setVisible] = useState(PAGE_SIZE);

  if (entries.length === 0) {
    return <p className="text-muted-foreground">Nenhuma movimentação ainda.</p>;
  }

  return (
    <div className="space-y-4">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Data</TableHead>
            <TableHead>Tipo</TableHead>
            <TableHead className="text-right">Valor</TableHead>
            <TableHead className="text-right">Saldo após</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {entries.slice(0, visible).map((entry) => (
            <TableRow key={entry.id}>
              <TableCell>
                {new Date(entry.created_at).toLocaleString("pt-BR")}
              </TableCell>
              <TableCell>
                {entry.entry_type === "debit" ? "Débito" : "Crédito"}
              </TableCell>
              <TableCell
                className={`text-right ${entry.entry_type === "debit" ? "text-red-600" : "text-green-600"}`}
              >
                {entry.entry_type === "debit" ? "-" : "+"}
                {formatBRL(entry.amount)}
              </TableCell>
              <TableCell className="text-right">{formatBRL(entry.balance_after)}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      {visible < entries.length && (
        <Button variant="outline" onClick={() => setVisible((v) => v + PAGE_SIZE)}>
          Carregar mais
        </Button>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Página do extrato (Server Component)**

```tsx
// web/app/accounts/[id]/statement/page.tsx
import Link from "next/link";
import { redirect } from "next/navigation";
import { apiFetch, apiErrorMessage } from "@/lib/api";
import type { LedgerEntry } from "@/lib/types";
import { StatementTable } from "./statement-table";

type Props = {
  params: Promise<{ id: string }>; // Next 15: params é Promise
};

export default async function StatementPage({ params }: Props) {
  const { id } = await params;

  const res = await apiFetch(`/accounts/${id}/statement`);
  if (res.status === 401) {
    redirect("/login");
  }

  return (
    <main className="mx-auto max-w-3xl space-y-6 p-6">
      <header className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Extrato</h1>
        <Link href="/dashboard" className="text-sm underline">
          Voltar ao dashboard
        </Link>
      </header>

      {res.ok ? (
        <StatementTable entries={(await res.json()) as LedgerEntry[]} />
      ) : (
        // 403 (conta de outro user), 404 etc. — exibe o message do contrato
        <p className="text-red-600">
          {await apiErrorMessage(res, "erro ao carregar extrato")}
        </p>
      )}
    </main>
  );
}
```

- [ ] **Step 3: Build + verificação manual**

Run: `npm run build` (em `web/`)
Expected: verde.

Manual: "Ver extrato" no dashboard → tabela mostra débito/crédito da transferência da Task 8. Acessar `/accounts/<id-de-outro-user>/statement` → mensagem de erro (403) sem quebrar a página. Mais de 10 lançamentos → botão "Carregar mais".

- [ ] **Step 4: Commit**

```bash
git add web/app/accounts/
git commit -m "feat: tela de extrato com carregar mais"
```

---

### Task 10: Fluxo de aceite completo (marco da Semana 4)

Nenhum arquivo novo — só verificação de ponta a ponta.

- [ ] **Step 1: Subir tudo**

```bash
docker-compose up -d postgres
# aplicar seed se ainda não aplicou:
docker exec -i project_pay_db psql -U pay_user -d project_pay < scripts/dev_seed.sql
go run ./cmd/api           # terminal 1
cd web ; npm run dev       # terminal 2
```

- [ ] **Step 2: Roteiro no navegador (`http://localhost:3000`)**

1. `/signup`: cadastrar user novo → redireciona pra `/login`.
2. Logar → `/dashboard`.
3. "Nova conta" 2x (duas contas pro mesmo user).
4. Creditar saldo na conta 1 (valor em centavos):
   `docker exec -it project_pay_db psql -U pay_user -d project_pay -c "UPDATE accounts SET balance = balance + 100000 WHERE account_number = '<numero-da-conta-1>';"`
   Refresh no dashboard → conta 1 mostra R$ 1.000,00.
5. "Transferir": origem conta 1, destino ID da conta 2, R$ 100,00 → saldos atualizam (R$ 900,00 / R$ 100,00).
6. "Ver extrato" da conta 1 → débito de R$ 100,00 com saldo após R$ 900,00; conta 2 → crédito.
7. Transferir R$ 10.000,00 (mais que o saldo) → erro "saldo insuficiente" no dialog.
8. "Sair" → `/dashboard` redireciona pra `/login`.
9. Acessar `/dashboard` sem cookie (aba anônima) → redireciona pra `/login`.

Expected: tudo conforme descrito. Critério do spec: cadastrar → logar → ver contas e saldos → transferir → ver extrato refletindo. Estético não importa; funcional importa.

- [ ] **Step 3: Commit final (se houve ajuste)**

```bash
git add -A
git commit -m "chore: ajustes finais do fluxo da semana 4"
```
