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
