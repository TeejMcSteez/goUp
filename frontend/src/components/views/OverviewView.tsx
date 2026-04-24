import Island from "../layout/Island";
import useServiceName from "../../hooks/useServiceName";
import useServiceData from "../../hooks/useServiceData";
import type { Service } from "../../types";

interface QuickStatsProps {
  services: Service[];
  error: string | null;
}

function QuickStats({ services, error }: QuickStatsProps) {
  const totalServices = services.length;
  const errorServices = services.filter((s) => s.error).length;
  const operationalServices = totalServices - errorServices;
  const uptimePercentage =
    totalServices > 0
      ? ((operationalServices / totalServices) * 100).toFixed(1)
      : 0;

  if (error) {
    return <p>Error loading stats: {error}</p>;
  }

  return (
    <div className="grid grid-cols-[repeat(auto-fit,minmax(200px,1fr))] gap-6 mb-8">
      <div className="text-center p-6 bg-elevated rounded-xl border border-border">
        <h3 className="m-0 mb-2 text-muted">Total Services</h3>
        <p className="text-[2rem] font-bold m-0 text-primary">{totalServices}</p>
      </div>
      <div className="text-center p-6 bg-elevated rounded-xl border border-border">
        <h3 className="m-0 mb-2 text-muted">Operational</h3>
        <p className="text-[2rem] font-bold m-0 text-success">{operationalServices}</p>
      </div>
      <div className="text-center p-6 bg-elevated rounded-xl border border-border">
        <h3 className="m-0 mb-2 text-muted">Errors</h3>
        <p className="text-[2rem] font-bold m-0 text-error">{errorServices}</p>
      </div>
      <div className="text-center p-6 bg-elevated rounded-xl border border-border">
        <h3 className="m-0 mb-2 text-muted">Uptime</h3>
        <p className="text-[2rem] font-bold m-0 text-primary">{uptimePercentage}%</p>
      </div>
    </div>
  );
}

interface MiniServiceGridProps {
  services: Service[];
  error: string | null;
}

function MiniServiceGrid({ services, error }: MiniServiceGridProps) {
  const { formatName } = useServiceName();

  if (error) {
    return <p>Error loading services: {error}</p>;
  }

  const errorServices = services.filter((s) => s.error);
  const displayServices = errorServices.length > 0 ? errorServices : services;

  return (
    <div className="grid grid-cols-[repeat(auto-fit,minmax(250px,1fr))] gap-4">
      {displayServices.map((service) => (
        <div
          key={service.name}
          className={`p-4 bg-elevated rounded-lg border border-l-4 ${
            service.error
              ? "border-error border-l-error"
              : "border-success border-l-success"
          }`}
        >
          <h4 className="m-0 mb-2">
            <a href={service.url} target="_blank" rel="noreferrer">
              {formatName(service.name)}
            </a>
          </h4>
          <p className={`m-0 text-sm ${service.error ? "text-error" : "text-success"}`}>
            {service.error ? "❌ Error" : "✅ Operational"}
          </p>
          <p className="mt-1 mb-0 text-sm text-muted">{service.response_time}</p>
        </div>
      ))}
    </div>
  );
}

export default function OverviewView() {
  const { data: services, error } = useServiceData();

  return (
    <div>
      <Island title="System Status">
        <QuickStats services={services ?? []} error={error} />
      </Island>
      <Island title="Service Overview">
        <MiniServiceGrid services={services ?? []} error={error} />
      </Island>
    </div>
  );
}
