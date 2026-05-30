import { useState } from "react";
import useServiceName from "../../../hooks/useServiceName";
import { ServiceCardProps } from "../../../types.ts";
import fallback from "../../../../static/goup.png";

export default function ServiceCard({ service }: ServiceCardProps) {
  const { url, name, description, response, response_time, data, error } =
    service;
  const { formatName } = useServiceName();
  const [showApiResponse, setShowApiResponse] = useState(false);
  const [showFullHttpResponse, setShowFullHttpResponse] = useState(false);

  const base = url.endsWith("/") ? url : url + "/";
  const faviconUrl = base + "favicon.ico";

  const isLongHttpResponse = error && response && response.length > 3;

  return (
    <div
      className={`bg-surface rounded-xl p-6 flex flex-col gap-3 transition-[transform,box-shadow,border-color] duration-200 hover:-translate-y-1.25 hover:shadow-lg border-l-4 border border-border ${
        error
          ? "hover:border-error border-l-error"
          : "hover:border-primary border-l-primary"
      }`}
    >
      <div className="flex items-start justify-between gap-2">
        <h3 className="m-0 text-fg font-semibold text-lg leading-tight flex items-center gap-2">
          <img
            src={faviconUrl}
            alt=""
            className="w-4 h-4 rounded-sm shrink-0"
            onError={(e) => {
              e.currentTarget.src = fallback;
            }}
          />
          <a href={url} target="_blank" rel="noreferrer">
            {formatName(name)}
          </a>
        </h3>
        <span
          className={`shrink-0 text-xs font-semibold px-2 py-1 rounded-full ${
            error ? "bg-error/15 text-error" : "bg-success/15 text-success"
          }`}
        >
          {error ? "Error" : "Operational"}
        </span>
      </div>

      <div className="h-px bg-border" />

      <div className={description ? "flex flex-wrap" : "hidden"}>
        <p className="text-xs">{description}</p>
      </div>

      <div className="flex flex-wrap gap-3">
        <div className="flex flex-col">
          <span className="text-[0.7rem] uppercase tracking-wider text-muted font-medium">
            Response Time
          </span>
          <span className="text-sm font-semibold text-primary">
            {response_time ?? "—"}
          </span>
        </div>
        <div className="flex flex-col">
          <span className="text-[0.7rem] uppercase tracking-wider text-muted font-medium">
            HTTP
          </span>
          {isLongHttpResponse ? (
            <button
              onClick={() => setShowFullHttpResponse(!showFullHttpResponse)}
              className="text-sm font-semibold text-error text-left p-0 border-none bg-transparent cursor-pointer hover:translate-y-0 hover:bg-transparent hover:border-none"
            >
              {showFullHttpResponse ? "Hide ▲" : "Show ▼"}
            </button>
          ) : (
            <span
              className={`text-sm font-semibold ${error ? "text-error" : "text-success"}`}
            >
              {response ?? "—"}
            </span>
          )}
        </div>
      </div>

      {isLongHttpResponse && showFullHttpResponse && (
        <div className="min-h-0 max-h-37.5 overflow-y-auto bg-error/10 border border-error/30 rounded-lg p-3 text-sm text-error wrap-break-words scrollbar-custom">
          {response}
        </div>
      )}

      <button
        onClick={() => setShowApiResponse(!showApiResponse)}
        className="flex items-center justify-between w-full mt-1 px-3 py-2 rounded-lg bg-elevated border border-border text-muted text-sm font-medium cursor-pointer hover:text-fg hover:border-primary hover:bg-hover hover:translate-y-0 transition-colors duration-200"
      >
        <span>API Response</span>
        <span className="text-xs">{showApiResponse ? "▲" : "▼"}</span>
      </button>

      {showApiResponse && (
        <div className="min-h-0 max-h-50 overflow-y-auto bg-hover p-4 rounded-lg text-sm text-muted whitespace-pre-wrap wrap-break-words scrollbar-custom">
          {data ?? "No API setup in configuration"}
        </div>
      )}
    </div>
  );
}
