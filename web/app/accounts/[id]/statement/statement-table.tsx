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
