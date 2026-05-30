import { useState } from "react";
import type { ServiceConfig } from "../../types";

interface RowData {
  key: string;
  Name: string;
  URL: string;
  Description: string;
  API_URL: string;
  Valid_Responses: string;
  Retry_Requests: string;
}

interface BulkEditPanelProps {
  services: Record<string, ServiceConfig>;
  onSave: (updates: { oldName: string; service: Partial<ServiceConfig> }[]) => Promise<void>;
  onCancel: () => void;
}

const inputClass =
  "w-full px-2 py-1 rounded border border-border bg-surface text-fg text-[0.85rem] transition-[border-color,box-shadow] duration-200 focus:outline-none focus:border-focus focus:shadow-[0_0_0_3px_rgba(56,189,248,0.15)] placeholder:text-muted placeholder:opacity-50";

const btnBase =
  "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

const cols = ["Name", "URL", "Description", "API URL", "Valid Responses", "Retries"] as const;

export default function BulkEditPanel({ services, onSave, onCancel }: BulkEditPanelProps) {
  const [rows, setRows] = useState<RowData[]>(() =>
    Object.entries(services).map(([key, svc]) => ({
      key,
      Name: svc.Name ?? "",
      URL: svc.URL ?? "",
      Description: svc.Description ?? "",
      API_URL: svc.API_URL ?? "",
      Valid_Responses: svc.Valid_Responses?.join(", ") ?? "",
      Retry_Requests: svc.Retry_Requests != null ? String(svc.Retry_Requests) : "",
    }))
  );
  const [saving, setSaving] = useState(false);

  const setCell =
    (idx: number, field: keyof Omit<RowData, "key">) =>
    (e: React.ChangeEvent<HTMLInputElement>) =>
      setRows((prev) => prev.map((r, i) => (i === idx ? { ...r, [field]: e.target.value } : r)));

  const handleSave = async () => {
    setSaving(true);
    const updates = rows.map((row) => {
      const validResponses = row.Valid_Responses
        ? row.Valid_Responses.split(",").map((s) => s.trim()).filter(Boolean)
        : null;
      return {
        oldName: row.key,
        service: {
          Name: row.Name,
          URL: row.URL,
          ...(row.Description && { Description: row.Description }),
          ...(row.API_URL && { API_URL: row.API_URL }),
          ...(validResponses?.length ? { Valid_Responses: validResponses.map(Number) } : { Valid_Responses: undefined }),
          ...(row.Retry_Requests !== "" ? { Retry_Requests: parseInt(row.Retry_Requests) } : { Retry_Requests: undefined }),
        } as Partial<ServiceConfig>,
      };
    });
    await onSave(updates);
    setSaving(false);
  };

  return (
    <div className="flex flex-col gap-4 p-3 sm:p-4 bg-elevated border border-border rounded-lg">
      {/* Column headers — desktop only */}
      <div className="hidden sm:grid sm:grid-cols-[1fr_2fr_2fr_2fr_1.5fr_0.6fr] gap-2 text-muted text-[0.8rem] font-medium pb-1 border-b border-border">
        {cols.map((c) => <span key={c}>{c}</span>)}
      </div>

      <div className="flex flex-col gap-3">
        {rows.map((row, idx) => (
          <div
            key={row.key}
            className="flex flex-col gap-2 p-3 bg-surface border border-border rounded-lg sm:p-0 sm:bg-transparent sm:border-x-0 sm:border-t-0 sm:rounded-none sm:grid sm:grid-cols-[1fr_2fr_2fr_2fr_1.5fr_0.6fr] sm:gap-2 sm:items-center sm:pb-2"
          >
            <div className="flex flex-col gap-1">
              <label className="sm:hidden text-muted text-[0.75rem] font-medium">Name</label>
              <input className={inputClass} value={row.Name} onChange={setCell(idx, "Name")} placeholder="Name" required />
            </div>
            <div className="flex flex-col gap-1">
              <label className="sm:hidden text-muted text-[0.75rem] font-medium">URL</label>
              <input className={inputClass} value={row.URL} onChange={setCell(idx, "URL")} placeholder="https://example.com" required />
            </div>
            <div className="flex flex-col gap-1">
              <label className="sm:hidden text-muted text-[0.75rem] font-medium">Description</label>
              <input className={inputClass} value={row.Description} onChange={setCell(idx, "Description")} placeholder="What this service does" />
            </div>
            <div className="flex flex-col gap-1">
              <label className="sm:hidden text-muted text-[0.75rem] font-medium">API URL</label>
              <input className={inputClass} value={row.API_URL} onChange={setCell(idx, "API_URL")} placeholder="https://api.example.com" />
            </div>
            <div className="flex flex-col gap-1">
              <label className="sm:hidden text-muted text-[0.75rem] font-medium">Valid Responses</label>
              <input className={inputClass} value={row.Valid_Responses} onChange={setCell(idx, "Valid_Responses")} placeholder="200, 201" />
            </div>
            <div className="flex flex-col gap-1">
              <label className="sm:hidden text-muted text-[0.75rem] font-medium">Retries</label>
              <input className={inputClass} type="number" min="0" value={row.Retry_Requests} onChange={setCell(idx, "Retry_Requests")} placeholder="3" />
            </div>
          </div>
        ))}
      </div>

      <div className="flex gap-2 flex-wrap">
        <button
          className={`${btnBase} border-primary text-primary hover:bg-primary/10`}
          onClick={handleSave}
          disabled={saving}
        >
          {saving ? "Saving…" : "Save All"}
        </button>
        <button className={btnBase} onClick={onCancel} disabled={saving}>
          Cancel
        </button>
      </div>
    </div>
  );
}
