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

  if (error) {
    return <p>Error loading stats: {error}</p>;
  }

  return (
    <div className="grid grid-cols-[repeat(auto-fit,minmax(min(200px,100%),1fr))] gap-6 mb-8">
      <div className="text-center p-6 bg-elevated rounded-xl border border-border">
        <h3 className="m-0 mb-2 text-muted">Total Services</h3>
        <p className="text-[2rem] font-bold m-0 text-primary">
          {totalServices}
        </p>
      </div>
      <div className="text-center p-6 bg-elevated rounded-xl border border-border">
        <h3 className="m-0 mb-2 text-muted">Operational</h3>
        <p className="text-[2rem] font-bold m-0 text-success">
          {operationalServices}
        </p>
      </div>
      <div className="text-center p-6 bg-elevated rounded-xl border border-border">
        <h3 className="m-0 mb-2 text-muted">Errors</h3>
        <p className="text-[2rem] font-bold m-0 text-error">{errorServices}</p>
      </div>
    </div>
  );
}

interface MiniServiceGridProps {
  services: Service[];
  error: string | null;
}

function parseMs(rt?: string): number | null {
  if (!rt) return null;
  const match = rt.match(/([\d.]+)\s*(ms|s)/i);
  if (!match) return null;
  const val = parseFloat(match[1]);
  return match[2].toLowerCase() === "s" ? val * 1000 : val;
}

function MiniServiceGrid({ services, error }: MiniServiceGridProps) {
  const { formatName } = useServiceName();

  if (error) {
    return <p>Error loading services: {error}</p>;
  }

  if (!services.length) {
    return <p className="text-muted">No services to display.</p>;
  }

  const errored = services.filter((s) => s.error);
  const operational = services.filter((s) => !s.error);

  const timed = services
    .map((s) => ({ service: s, ms: parseMs(s.response_time) }))
    .filter((x): x is { service: Service; ms: number } => x.ms !== null)
    .sort((a, b) => a.ms - b.ms);

  const fastest = timed[0];
  const slowest = timed[timed.length - 1];
  const hasDistinctSlowest =
    slowest && fastest && slowest.service.name !== fastest.service.name;

  return (
    <div className="flex flex-col gap-4">
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div
          className={`p-5 rounded-xl border ${
            errored.length > 0
              ? "bg-error/10 border-error/30"
              : "bg-success/10 border-success/30"
          }`}
        >
          <span className="text-xs uppercase tracking-wider font-medium text-muted">
            System Health
          </span>
          <p
            className={`text-2xl font-bold mt-2 mb-1 ${
              errored.length > 0 ? "text-error" : "text-success"
            }`}
          >
            {errored.length > 0
              ? `${errored.length} Error${errored.length > 1 ? "s" : ""}`
              : "All Clear"}
          </p>
          <p className="text-sm text-muted m-0">
            {operational.length}/{services.length} operational
          </p>
        </div>

        {fastest && (
          <div className="p-5 rounded-xl border bg-elevated border-border">
            <span className="text-xs uppercase tracking-wider font-medium text-muted">
              Fastest Response
            </span>
            <p className="text-lg font-bold mt-2 mb-1 text-fg">
              <a
                href={fastest.service.url}
                target="_blank"
                rel="noreferrer"
                className="hover:text-primary transition-colors"
              >
                {formatName(fastest.service.name)}
              </a>
            </p>
            <p className="text-sm font-semibold text-primary m-0">
              {fastest.service.response_time}
            </p>
          </div>
        )}

        {hasDistinctSlowest && (
          <div className="p-5 rounded-xl border bg-elevated border-border">
            <span className="text-xs uppercase tracking-wider font-medium text-muted">
              Slowest Response
            </span>
            <p className="text-lg font-bold mt-2 mb-1 text-fg">
              <a
                href={slowest.service.url}
                target="_blank"
                rel="noreferrer"
                className="hover:text-secondary transition-colors"
              >
                {formatName(slowest.service.name)}
              </a>
            </p>
            <p className="text-sm font-semibold text-secondary m-0">
              {slowest.service.response_time}
            </p>
          </div>
        )}
      </div>

      {errored.length > 0 && (
        <div className="p-5 rounded-xl border border-error/30 bg-error/5">
          <span className="text-xs uppercase tracking-wider font-medium text-muted mb-3 block">
            Services with Errors
          </span>
          <div className="flex flex-col">
            {errored.map((s) => (
              <div
                key={s.name}
                className="flex items-center justify-between py-3 border-b border-error/15 last:border-0"
              >
                <a
                  href={s.url}
                  target="_blank"
                  rel="noreferrer"
                  className="font-semibold text-fg hover:text-primary transition-colors"
                >
                  {formatName(s.name)}
                </a>
                <div className="flex items-center gap-3">
                  {s.response_time && (
                    <span className="text-xs text-muted">
                      {s.response_time}
                    </span>
                  )}
                  <a
                    className="cursor-pointer text-xs text-error bg-error/15 px-2 py-1 rounded-full font-medium"
                    onClick={() =>
                      window.dispatchEvent(
                        new CustomEvent("goup:navigate", {
                          detail: "services",
                        }),
                      )
                    }
                  >
                    Error
                  </a>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
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
