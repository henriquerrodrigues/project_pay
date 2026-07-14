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
