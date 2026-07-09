import FormPanel, { type FieldDef } from "./FormPanel";
import type { SMTPConfig, SMTPPanelProps } from "../../types";

interface SMTPFormData {
  [key: string]: string;
  Email: string;
  App_Password: string;
  SMTPServer: string;
  Backoff_Period: string;
}

const EMPTY: SMTPFormData = {
  Email: "",
  App_Password: "",
  SMTPServer: "",
  Backoff_Period: "",
};

const FIELDS: FieldDef<SMTPFormData>[] = [
  { key: "SMTPServer", label: "SMTP Server", placeholder: "smtp.example.com:587" },
  { key: "Email", label: "Email", type: "email", placeholder: "you@example.com" },
  { key: "App_Password", label: "App Password", type: "password", placeholder: "••••••••" },
  { key: "Backoff_Period", label: "Backoff Period", placeholder: "5m" },
];

function toForm(smtp?: SMTPConfig): SMTPFormData {
  return {
    Email: smtp?.Email ?? "",
    App_Password: smtp?.App_Password ?? "",
    SMTPServer: smtp?.SMTPServer ?? "",
    Backoff_Period: smtp?.Backoff_Period ?? "",
  };
}

export default function SMTPPanel({ smtp, onRefresh }: SMTPPanelProps) {
  return (
    <FormPanel
      endpoint="/api/config/smtp"
      entityLabel="SMTP"
      fields={FIELDS}
      initial={toForm(smtp)}
      empty={EMPTY}
      onRefresh={onRefresh}
    />
  );
}
