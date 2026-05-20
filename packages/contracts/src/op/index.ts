export interface OpVault {
  id: string;
  name: string;
}

export interface OpItem {
  id: string;
  title: string;
}

export interface OpUnavailableError extends Error {
  code: "op_unavailable";
}
