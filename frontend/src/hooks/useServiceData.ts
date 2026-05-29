import { useCallback } from "react";
import usePolling from "./usePolling";
import type { Service } from "../types";
import { POLL_RATE } from "../constants";

export default function useServiceData() {
  const fetchServices = useCallback(async (): Promise<Service[]> => {
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
  }, []);

  return usePolling(fetchServices, POLL_RATE);
}
