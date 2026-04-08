import { useState } from "react";
import { useConfigData } from "../hooks/useConfigData";
import ServicesPanel from "./config/ServicesPanel";
import MQTTPanel from "./config/MQTTPanel";
import WebhookPanel from "./config/WebhookPanel";
import DatabasePanel from "./config/DatabasePanel";

const SECTIONS = [
  { id: "services", label: "Services" },
  { id: "mqtt", label: "MQTT" },
  { id: "webhook", label: "Webhook" },
  { id: "database", label: "Database" },
];

export default function ConfigEditor() {
  const { config, loading, error, refresh } = useConfigData();
  const [activeSection, setActiveSection] = useState("services");

  if (loading) return <p className="config-loading">Loading configuration…</p>;
  if (error) return <p className="config-status config-status--error">{error}</p>;

  return (
    <div className="config-editor">
      <div className="config-section-tabs">
        {SECTIONS.map((s) => (
          <button
            key={s.id}
            className={`config-section-tab ${activeSection === s.id ? "active" : ""}`}
            onClick={() => setActiveSection(s.id)}
          >
            {s.label}
          </button>
        ))}
      </div>
      <div className="config-section-content">
        {activeSection === "services" && <ServicesPanel services={config?.services} onRefresh={refresh} />}
        {activeSection === "mqtt" && <MQTTPanel mqtt={config?.mqtt} onRefresh={refresh} />}
        {activeSection === "webhook" && <WebhookPanel webhook={config?.webhook} onRefresh={refresh} />}
        {activeSection === "database" && <DatabasePanel />}
      </div>
    </div>
  );
}
