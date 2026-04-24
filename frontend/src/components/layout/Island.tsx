import type { ReactNode } from "react";

interface IslandProps {
  title?: string;
  children: ReactNode;
  headerAction?: ReactNode;
}

export default function Island({ title, children, headerAction }: IslandProps) {
  return (
    <div className="bg-surface rounded-2xl p-4 lg:p-6 border border-border shadow-md mb-6 lg:mb-8 transition-[transform,box-shadow] duration-200 hover:-translate-y-0.5 hover:shadow-lg">
      {title && (
        <div className="flex flex-col items-start gap-2 sm:flex-row sm:justify-between sm:items-center mb-6 pb-4 border-b border-border">
          <h2 className="m-0 text-primary text-xl sm:text-2xl font-semibold">{title}</h2>
          {headerAction && <div className="flex gap-2">{headerAction}</div>}
        </div>
      )}
      <div>{children}</div>
    </div>
  );
}
