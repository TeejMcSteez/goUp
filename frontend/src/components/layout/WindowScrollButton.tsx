import { useState, useEffect } from "react";

export default function WindowScrollButton() {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const onScroll = () => setVisible(window.scrollY > 200);
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  if (!visible) return null;

  return (
    <button
      onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}
      className="fixed bottom-4 left-4 z-50 px-3 py-2 rounded-lg bg-elevated border border-border text-muted text-sm hover:text-fg hover:border-primary hover:translate-y-0 transition-colors"
    >
      ↑ Top
    </button>
  );
}
