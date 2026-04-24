import { useEffect } from "react";

interface ConfirmModalProps {
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
}

export default function ConfirmModal({ message, onConfirm, onCancel }: ConfirmModalProps) {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === "Escape") onCancel(); };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onCancel]);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={onCancel}
    >
      <div
        className="bg-surface border border-border rounded-xl shadow-xl p-6 flex flex-col gap-4 min-w-[280px] max-w-sm w-full mx-4"
        onClick={(e) => e.stopPropagation()}
      >
        <p className="text-fg text-[0.95rem]">{message}</p>
        <div className="flex gap-2 justify-end">
          <button
            className="px-4 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated"
            onClick={onCancel}
          >
            Cancel
          </button>
          <button
            className="px-4 py-2 rounded-lg border border-error bg-surface text-error text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-error/10"
            onClick={onConfirm}
          >
            Remove
          </button>
        </div>
      </div>
    </div>
  );
}
