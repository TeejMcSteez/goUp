import useLiveQuery from "./useLiveQuery";
import type { TlsStatus } from "../types";

async function fetchTlsData(): Promise<TlsStatus[]> {
  const res = await fetch("/api/tls");
  if (!res.ok) {
    throw new Error(`Server error: ${res.status}`);
  }
  const data: unknown = await res.json();
  return (data as TlsStatus[]) || [];
}

export default function useTlsData() {
  return useLiveQuery(["tls"], fetchTlsData);
}
