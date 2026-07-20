import { useState, useEffect } from "react";

const STORAGE_KEY = "goUp_prettify_names";
const EVENT = "goUp:prettify-names-changed";

function normalizeName(name: string): string {
  return name
    .replace(/^https?:\/\//i, "")
    .replace(/_/g, " ")
    .trim();
}

interface ServiceNameHook {
  formatName: (name: string) => string;
  prettify: boolean;
  toggle: () => void;
}

export default function useServiceName(): ServiceNameHook {
  const [prettify, setPrettify] = useState(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored !== null ? (JSON.parse(stored) as boolean) : true;
  });

  useEffect(() => {
    const handler = (e: Event) => setPrettify((e as CustomEvent<boolean>).detail);
    window.addEventListener(EVENT, handler);
    return () => window.removeEventListener(EVENT, handler);
  }, []);

  const toggle = () => {
    setPrettify((prev) => {
      const next = !prev;
      localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
      window.dispatchEvent(new CustomEvent<boolean>(EVENT, { detail: next }));
      return next;
    });
  };

  const formatName = (name: string) => (prettify ? normalizeName(name) : name);

  return { formatName, prettify, toggle };
}
