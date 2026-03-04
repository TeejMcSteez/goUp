import { useState, useEffect, useCallback } from "react";

const STORAGE_KEY = "goUp_prettify_names";
const EVENT = "goUp:prettify-names-changed";

function normalizeName(name) {
  if (!name) return name;
  return name
    .replace(/^https?:\/\//i, "") // strip URL prefix
    .replace(/_/g, " ")           // underscores → spaces
    .trim();
}

export default function useServiceName() {
  const [prettify, setPrettify] = useState(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored !== null ? JSON.parse(stored) : true;
  });

  useEffect(() => {
    const handler = (e) => setPrettify(e.detail);
    window.addEventListener(EVENT, handler);
    return () => window.removeEventListener(EVENT, handler);
  }, []);

  const toggle = useCallback(() => {
    setPrettify((prev) => {
      const next = !prev;
      localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
      window.dispatchEvent(new CustomEvent(EVENT, { detail: next }));
      return next;
    });
  }, []);

  const formatName = useCallback(
    (name) => (prettify ? normalizeName(name) : name),
    [prettify]
  );

  return { formatName, prettify, toggle };
}
