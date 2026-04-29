import { useState } from "react";
import { useConfigData } from "../hooks/useConfigData";
import ServicesPanel from "./config/ServicesPanel";
import MQTTPanel from "./config/MQTTPanel";
import WebhookPanel from "./config/WebhookPanel";
import DatabasePanel from "./config/DatabasePanel";

interface Section {
  id: string;
  label: string;
}

const SECTIONS: Section[] = [
  { id: "services", label: "Services" },
  { id: "mqtt", label: "MQTT" },
  { id: "webhook", label: "Webhook" },
  { id: "database", label: "Database" },
];

export default function ConfigEditor() {
  const { config, loading, error, refresh } = useConfigData();
  const [activeSection, setActiveSection] = useState("services");

  if (loading) return <p className="text-muted text-[0.9rem]">Loading configuration…</p>;
  if (error) return (
    <p className="text-[0.9rem] m-0 px-4 py-2 rounded-lg text-error bg-error/10 border border-error/30">
      {error}
    </p>
  );

  return (
    <div className="w-full min-w-0">
      <div className="flex flex-wrap gap-1 mb-6 border-b border-border pb-2">
        {SECTIONS.map((s) => (
          <button
            key={s.id}
            className={`px-3 sm:px-6 py-2 border-none text-[0.9rem] font-medium cursor-pointer rounded-lg transition-all duration-200 hover:text-fg hover:bg-elevated hover:translate-y-0 ${
              activeSection === s.id
                ? "text-primary bg-elevated"
                : "bg-transparent text-muted"
            }`}
            onClick={() => setActiveSection(s.id)}
          >
            {s.label}
          </button>
        ))}
      </div>
      <div>
        {activeSection === "services" && <ServicesPanel services={config?.services} onRefresh={refresh} />}
        {activeSection === "mqtt" && <MQTTPanel mqtt={config?.mqtt} onRefresh={refresh} />}
        {activeSection === "webhook" && <WebhookPanel webhook={config?.webhook} onRefresh={refresh} />}
        {activeSection === "database" && <DatabasePanel />}
      </div>
    </div>
  );
}
