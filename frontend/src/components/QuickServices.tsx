import { useState, useEffect } from "react";
import type { Service } from "../types";

export default function QuickServices() {
  const [qs, setQs] = useState<Service[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const getDownServices = async () => {
      try {
        const res = await fetch("/api/status");
        if (!res.ok) {
          throw new Error(`Server error: ${res.status}`);
        }
        const data: unknown = await res.json();
        if (Array.isArray(data)) {
          setQs(data as Service[]);
        } else if (data === null) {
          setQs([]);
        } else {
          console.error(`Expected array from /api/status, got: ${String(data)}`);
        }
        setError(null);
      } catch (err) {
        setError((err as Error).message);
        console.error("Error fetching quick service data:", err);
      }
    };

    getDownServices();
    const intervalId = setInterval(getDownServices, 5000);
    return () => clearInterval(intervalId);
  }, []);

  const content =
    qs.length === 0 ? (
      <a
        href="#cards"
        className="no-underline text-center text-lg text-fg px-6 py-2 rounded-lg transition-colors duration-200 hover:bg-hover"
      >
        All Systems Operational ✅
      </a>
    ) : (
      <a
        href="#cards"
        className="no-underline text-center text-lg text-fg px-6 py-2 rounded-lg transition-colors duration-200 hover:bg-hover"
      >
        {qs.length} Errors Detected ❌
      </a>
    );

  return (
    <div className="flex flex-row items-center justify-center text-center p-4 border-b border-border bg-surface">
      {error ? (
        <span className="text-error text-xs flex flex-row justify-center items-center">
          <p className="text-error text-xs animate-ping mr-2">O</p> Status
          unavailable: {error}
        </span>
      ) : (
        content
      )}
    </div>
  );
}
