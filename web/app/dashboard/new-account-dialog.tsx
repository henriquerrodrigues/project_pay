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
