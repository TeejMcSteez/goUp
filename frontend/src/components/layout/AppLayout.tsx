import {
  useState,
  useEffect,
  useCallback,
  lazy,
  Suspense,
  Component,
  type ReactNode,
  type ComponentType,
} from "react";
import Header from "../Header";
import Navigation from "./Navigation";

const nullModule: { default: ComponentType } = { default: () => null };
const OverviewView = lazy(() =>
  import("../views/OverviewView").catch(() => nullModule),
);
const ServicesView = lazy(() =>
  import("../views/ServicesView").catch(() => nullModule),
);
const AnalyticsView = lazy(() =>
  import("../views/AnalyticsView").catch(() => nullModule),
);
const SettingsView = lazy(() =>
  import("../views/SettingsView").catch(() => nullModule),
);

const VIEWS: Record<string, ComponentType> = {
  overview: OverviewView,
  services: ServicesView,
  analytics: AnalyticsView,
  settings: SettingsView,
};

const KEY_MAP: Record<string, string> = {
  "1": "overview",
  "2": "services",
  "3": "analytics",
  "4": "settings",
};

const SPINNER = (
  <div className="flex items-center justify-center py-24">
    <div className="w-6 h-6 rounded-full border-2 border-primary/20 border-t-primary animate-spin" />
  </div>
);

interface ServerDownBannerProps {
  networkOnline: boolean;
}

function ServerDownBanner({ networkOnline }: ServerDownBannerProps) {
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

function WindowScrollButton() {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const onScroll = () => setVisible(window.scrollY > 200);
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  if (!visible) return null;

  return (
    <button
      onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}
      className="fixed bottom-4 left-4 z-50 px-3 py-2 rounded-lg bg-elevated border border-border text-muted text-sm hover:text-fg hover:border-primary hover:translate-y-0 transition-colors"
    >
      ↑ Top
    </button>
  );
}

interface ChunkErrorBoundaryProps {
  resetKey: boolean;
  children: ReactNode;
}

interface ChunkErrorBoundaryState {
  crashed: boolean;
}

class ChunkErrorBoundary extends Component<
  ChunkErrorBoundaryProps,
  ChunkErrorBoundaryState
> {
  constructor(props: ChunkErrorBoundaryProps) {
    super(props);
    this.state = { crashed: false };
  }

  static getDerivedStateFromError(): ChunkErrorBoundaryState {
    return { crashed: true };
  }

  componentDidUpdate(prevProps: ChunkErrorBoundaryProps) {
    if (prevProps.resetKey !== this.props.resetKey && this.state.crashed) {
      this.setState({ crashed: false });
    }
  }

  render() {
    if (this.state.crashed) return null;
    return this.props.children;
  }
}

export default function AppLayout() {
  const [activeTab, setActiveTab] = useState(() => {
    return localStorage.getItem("activeTab") ?? "overview";
  });
  const [serverDown, setServerDown] = useState(false);
  const [networkOnline, setNetworkOnline] = useState(navigator.onLine);

  const handleTabChange = useCallback((tab: string) => {
    localStorage.setItem("activeTab", tab);
    setActiveTab(tab);
  }, []);

  useEffect(() => {
    const check = async () => {
      setNetworkOnline(navigator.onLine);
      try {
        const res = await fetch("/api", { signal: AbortSignal.timeout(4000) });
        setServerDown(!res.ok);
      } catch {
        setServerDown(true);
      }
    };

    check();
    const id = setInterval(check, 10000);
    return () => clearInterval(id);
  }, []);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key in KEY_MAP) {
        e.preventDefault();
        handleTabChange(KEY_MAP[e.key]);
      }
      if (e.key === "ArrowLeft" || e.key === "ArrowRight") {
        if (activeTab === "settings") return;
        const tag = (document.activeElement as HTMLElement)?.tagName;
        const editable = (document.activeElement as HTMLElement)
          ?.isContentEditable;
        if (
          tag === "INPUT" ||
          tag === "TEXTAREA" ||
          tag === "SELECT" ||
          editable
        )
          return;
        const views = Object.values(KEY_MAP);
        const current = views.indexOf(activeTab);
        if (e.key === "ArrowLeft") {
          handleTabChange(views[(current - 1 + views.length) % views.length]);
        } else {
          handleTabChange(views[(current + 1) % views.length]);
        }
      }
    };

    const handleNavigate = (e: Event) =>
      handleTabChange((e as CustomEvent<string>).detail);

    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("goup:navigate", handleNavigate);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("goup:navigate", handleNavigate);
    };
  }, [handleTabChange, activeTab]);

  const ActiveView = VIEWS[activeTab] ?? OverviewView;

  return (
    <div className="flex flex-col min-h-screen w-full">
      <Header />
      <Navigation activeTab={activeTab} onTabChange={handleTabChange} />
      <main
        className="flex-1 max-w-350 w-full mx-auto p-4 md:p-6 xl:p-8 animate-fade-in"
        role="tabpanel"
        id={`${activeTab}-panel`}
        aria-labelledby={`${activeTab}-tab`}
      >
        <ChunkErrorBoundary resetKey={serverDown}>
          <Suspense fallback={SPINNER}>
            {serverDown ? (
              <ServerDownBanner networkOnline={networkOnline} />
            ) : (
              <ActiveView />
            )}
          </Suspense>
        </ChunkErrorBoundary>
      </main>
      <WindowScrollButton />
    </div>
  );
}
