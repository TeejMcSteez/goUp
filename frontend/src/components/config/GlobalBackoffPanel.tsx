import FormPanel, { type FieldDef } from "./FormPanel";
import type { GlobalBackoffPanelProps } from "../../types";

interface BackoffFormData {
  [key: string]: string;
  backoff_period: string;
}

const EMPTY: BackoffFormData = { backoff_period: "" };

const FIELDS: FieldDef<BackoffFormData>[] = [
  { key: "backoff_period", label: "Backoff Period", placeholder: "5m" },
];

const DESCRIPTION = (
  <p className="text-[0.8rem] text-muted m-0">
    Default backoff applied to any notification trigger that doesn't
    set its own backoff period. Leave blank to disable backoff by
    default.
  </p>
);

export default function GlobalBackoffPanel({
  backoffPeriod,
  onRefresh,
}: GlobalBackoffPanelProps) {
  return (
    <FormPanel
      endpoint="/api/config/backoff"
      entityLabel="Global backoff"
      fields={FIELDS}
      initial={{ backoff_period: backoffPeriod ?? "" }}
      empty={EMPTY}
      description={DESCRIPTION}
      serialize={(f) => ({ backoff_period: f.backoff_period })}
      clear={() =>
        fetch("/api/config/backoff", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ backoff_period: "" }),
        })
      }
      onRefresh={onRefresh}
    />
  );
}
