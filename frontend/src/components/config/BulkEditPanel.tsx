import { useState } from "react";
import type { ServiceConfig } from "../../types";

interface RowData {
  key: string;
  Name: string;
  URL: string;
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

const cols = ["Name", "URL", "API URL", "Valid Responses", "Retries"] as const;

export default function BulkEditPanel({ services, onSave, onCancel }: BulkEditPanelProps) {
  const [rows, setRows] = useState<RowData[]>(() =>
    Object.entries(services).map(([key, svc]) => ({
      key,
      Name: svc.Name ?? "",
      URL: svc.URL ?? "",
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
    <div className="flex flex-col gap-4 p-4 bg-elevated border border-border rounded-lg">
      <div className="overflow-x-auto">
        <table className="w-full border-collapse text-[0.85rem]">
          <thead>
            <tr>
              {cols.map((c) => (
                <th key={c} className="text-left text-muted font-medium pb-2 pr-3 whitespace-nowrap">
                  {c}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {rows.map((row, idx) => (
              <tr key={row.key}>
                <td className="py-2 pr-3">
                  <input className={inputClass} value={row.Name} onChange={setCell(idx, "Name")} placeholder="Name" required />
                </td>
                <td className="py-2 pr-3">
                  <input className={inputClass} value={row.URL} onChange={setCell(idx, "URL")} placeholder="https://example.com" required />
                </td>
                <td className="py-2 pr-3">
                  <input className={inputClass} value={row.API_URL} onChange={setCell(idx, "API_URL")} placeholder="https://api.example.com" />
                </td>
                <td className="py-2 pr-3">
                  <input className={inputClass} value={row.Valid_Responses} onChange={setCell(idx, "Valid_Responses")} placeholder="200, 201" />
                </td>
                <td className="py-2 pr-3">
                  <input className={inputClass} type="number" min="0" value={row.Retry_Requests} onChange={setCell(idx, "Retry_Requests")} placeholder="3" />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
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
