import { useState } from "react";
import ServiceForm from "./ServiceForm";
import BulkEditPanel from "./BulkEditPanel";
import ConfirmModal from "./ConfirmModal";
import StatusMessage from "./StatusMessage";
import type { ServiceConfig, StatusMessage as StatusMsg } from "../../types";

interface ServiceFormData {
  Name: string;
  URL: string;
  Description: string;
  API_URL: string;
  Valid_Responses: string;
  Retry_Requests: string;
}

const EMPTY_FORM: ServiceFormData = {
  Name: "",
  URL: "",
  Description: "",
  API_URL: "",
  Valid_Responses: "",
  Retry_Requests: "",
};

function svcToForm(svc: ServiceConfig): ServiceFormData {
  return {
    Name: svc.Name ?? "",
    URL: svc.URL ?? "",
    Description: svc.Description ?? "",
    API_URL: svc.API_URL ?? "",
    Valid_Responses: svc.Valid_Responses?.join(", ") ?? "",
    Retry_Requests:
      svc.Retry_Requests != null ? String(svc.Retry_Requests) : "",
  };
}

const btnBase =
  "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

type Mode =
  | { type: "list" }
  | { type: "add" }
  | { type: "edit"; key: string }
  | { type: "bulk" };

interface ServicesPanelProps {
  services?: Record<string, ServiceConfig>;
  onRefresh: () => void;
}

export default function ServicesPanel({
  services,
  onRefresh,
}: ServicesPanelProps) {
  const [mode, setMode] = useState<Mode>({ type: "list" });
  const [status, setStatus] = useState<StatusMsg | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<ServiceConfig | null>(
    null,
  );

  const handleAdd = async (payload: Partial<ServiceConfig>) => {
    const res = await fetch("/api/config/service", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (res.ok) {
      setStatus({ text: "Service added.", error: false });
      setMode({ type: "list" });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleUpdate = async (
    oldName: string,
    payload: Partial<ServiceConfig>,
  ) => {
    const res = await fetch("/api/config/service", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ old_name: oldName, service: payload }),
    });
    if (res.ok) {
      setStatus({ text: "Service updated.", error: false });
      setMode({ type: "list" });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleBulkSave = async (
    updates: { oldName: string; service: Partial<ServiceConfig> }[],
  ) => {
    const results = await Promise.all(
      updates.map(({ oldName, service }) =>
        fetch("/api/config/service", {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ old_name: oldName, service }),
        }),
      ),
    );
    const failed = results.filter((r) => !r.ok);
    if (failed.length === 0) {
      setStatus({ text: `${updates.length} services updated.`, error: false });
      setMode({ type: "list" });
      onRefresh();
    } else {
      const errText = await failed[0].text();
      setStatus({
        text: `${failed.length} update(s) failed: ${errText}`,
        error: true,
      });
      onRefresh();
    }
  };

  const handleDelete = async (svc: ServiceConfig) => {
    const res = await fetch("/api/config/service", {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(svc),
    });
    if (res.ok) {
      setStatus({ text: "Service removed.", error: false });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const serviceList = Object.entries(services ?? {});

  return (
    <div className="flex flex-col gap-4">
      <StatusMessage message={status?.text} isError={status?.error} />

      {serviceList.length === 0 ? (
        <p className="text-muted text-[0.9rem]">No services configured.</p>
      ) : (
        <ul className="list-none m-0 p-0 flex flex-col gap-2">
          {serviceList.map(([key, svc]) => (
            <li
              key={key}
              className="flex items-center justify-between px-4 py-2 bg-elevated border border-border rounded-lg gap-4"
            >
              {mode.type === "edit" && mode.key === key ? (
                <ServiceForm
                  key={key}
                  initial={svcToForm(svc)}
                  onSubmit={(payload) => handleUpdate(key, payload)}
                  onCancel={() => setMode({ type: "list" })}
                  submitLabel="Save"
                />
              ) : (
                <>
                  <div className="flex flex-col gap-0.5 min-w-0">
                    <span className="font-semibold text-fg text-[0.95rem]">
                      {svc.Name ?? key}
                    </span>
                    <span className="text-[0.8rem] text-muted overflow-hidden text-ellipsis whitespace-nowrap">
                      {svc.URL}
                    </span>
                    {svc.Valid_Responses && svc.Valid_Responses.length > 0 && (
                      <span className="text-[0.8rem] text-muted overflow-hidden text-ellipsis whitespace-nowrap">
                        Valid: {svc.Valid_Responses.join(", ")}
                      </span>
                    )}
                  </div>
                  <div className="flex gap-2 flex-wrap">
                    <button
                      className={btnBase}
                      onClick={() => setMode({ type: "edit", key })}
                    >
                      Edit
                    </button>
                    <button
                      className={`${btnBase} border-error text-error hover:bg-error/10`}
                      onClick={() => setConfirmDelete(svc)}
                    >
                      Remove
                    </button>
                  </div>
                </>
              )}
            </li>
          ))}
        </ul>
      )}

      {mode.type === "bulk" ? (
        <BulkEditPanel
          services={services ?? {}}
          onSave={handleBulkSave}
          onCancel={() => setMode({ type: "list" })}
        />
      ) : mode.type === "add" ? (
        <ServiceForm
          key="add"
          initial={EMPTY_FORM}
          onSubmit={handleAdd}
          onCancel={() => setMode({ type: "list" })}
          submitLabel="Add"
        />
      ) : (
        mode.type === "list" && (
          <div className="flex gap-2 flex-wrap">
            <button
              className={`${btnBase} border-primary text-primary hover:bg-primary/10`}
              onClick={() => setMode({ type: "add" })}
            >
              + Add Service
            </button>
            {serviceList.length > 1 && (
              <button
                className={`${btnBase} border-primary text-primary hover:bg-primary/10`}
                onClick={() => setMode({ type: "bulk" })}
              >
                Bulk Edit
              </button>
            )}
          </div>
        )
      )}

      {confirmDelete && (
        <ConfirmModal
          message={`Remove "${confirmDelete.Name}"?`}
          onConfirm={() => {
            handleDelete(confirmDelete);
            setConfirmDelete(null);
          }}
          onCancel={() => setConfirmDelete(null)}
        />
      )}
    </div>
  );
}
