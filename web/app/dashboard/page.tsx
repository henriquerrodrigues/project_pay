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
