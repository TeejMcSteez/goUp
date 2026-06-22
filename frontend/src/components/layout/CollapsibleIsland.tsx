import { useState } from "react";
import type { ReactNode } from "react";

interface CollapsibleIslandProps {
  title: string;
  children: ReactNode;
  headerAction?: ReactNode;
  defaultOpen?: boolean;
}

export default function CollapsibleIsland({
  title,
  children,
  headerAction,
  defaultOpen = true,
}: CollapsibleIslandProps) {
  const [open, setOpen] = useState(defaultOpen);

  return (
    <div className="bg-surface rounded-2xl border border-border shadow-md mb-6 lg:mb-8 transition-[transform,box-shadow] duration-200 hover:-translate-y-0.5 hover:shadow-lg">
      <div
        className="flex flex-col items-start gap-2 sm:flex-row sm:justify-between sm:items-center p-4 lg:p-6 cursor-pointer select-none"
        onClick={() => setOpen((o) => !o)}
      >
        <div className="flex items-center gap-2">
          <svg
            className={`w-4 h-4 text-muted shrink-0 transition-transform duration-200 ${open ? "rotate-90" : "rotate-0"}`}
            viewBox="0 0 16 16"
            fill="currentColor"
          >
            <path d="M6 4l4 4-4 4V4z" />
          </svg>
          <h2 className="m-0 text-primary text-xl sm:text-2xl font-semibold">{title}</h2>
        </div>
        {headerAction && (
          <div className="flex gap-2" onClick={(e) => e.stopPropagation()}>
            {headerAction}
          </div>
        )}
      </div>

      <div className={`grid transition-[grid-template-rows] duration-200 ${open ? "grid-rows-[1fr]" : "grid-rows-[0fr]"}`}>
        <div className="overflow-hidden">
          <div className="px-4 lg:px-6 pb-4 lg:pb-6 border-t border-border pt-4">
            {children}
          </div>
        </div>
      </div>
    </div>
  );
}
