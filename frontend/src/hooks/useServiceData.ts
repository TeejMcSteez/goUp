import useLiveQuery from "./useLiveQuery";
import type { Service } from "../types";

async function fetchServices(): Promise<Service[]> {
  const res = await fetch("/api");
  if (!res.ok) {
    throw new Error(`Server error: ${res.status}`);
  }
  const data: unknown = await res.json();
  if (data && Array.isArray(data)) {
    return data as Service[];
  } else {
    console.error("Expected array from /api, got:", data);
    return [];
  }
}

export default function useServiceData() {
  return useLiveQuery(["services"], fetchServices);
}
