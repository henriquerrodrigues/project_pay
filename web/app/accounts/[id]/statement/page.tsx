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
