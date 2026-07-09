interface ServerDownBannerProps {
  networkOnline: boolean;
}

export default function ServerDownBanner({ networkOnline }: ServerDownBannerProps) {
  return (
    <div className="flex flex-col items-center justify-center py-24 gap-6 text-center">
      <div className="flex flex-col items-center gap-2">
        <div className="w-12 h-12 rounded-full bg-error/10 border border-error/30 flex items-center justify-center text-2xl">
          ✕
        </div>
        <h2 className="text-xl font-semibold text-fg m-0">
          Server Unreachable
        </h2>
        <p className="text-muted text-sm m-0">
          GoUp is not responding. It may be down or restarting.
        </p>
      </div>
      <div className="flex flex-col gap-2 text-sm">
        <div className="flex items-center gap-2">
          <span
            className={`w-2 h-2 rounded-full ${networkOnline ? "bg-success" : "bg-error"}`}
          />
          <span className="text-muted">
            Network:{" "}
            <span className={networkOnline ? "text-success" : "text-error"}>
              {networkOnline ? "Online" : "Offline"}
            </span>
          </span>
        </div>
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-error" />
          <span className="text-muted">
            Server: <span className="text-error">Unreachable</span>
          </span>
        </div>
      </div>
      <p className="text-muted text-xs m-0">Retrying automatically…</p>
    </div>
  );
}
