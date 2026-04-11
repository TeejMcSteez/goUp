import { useState, useEffect, lazy, Suspense, Component } from "react";
import Header from "../Header.jsx";
import Navigation from "./Navigation.jsx";

const nullModule = { default: () => null };
const OverviewView = lazy(() => import("../views/OverviewView.jsx").catch(() => nullModule));
const ServicesView = lazy(() => import("../views/ServicesView.jsx").catch(() => nullModule));
const AnalyticsView = lazy(() => import("../views/AnalyticsView.jsx").catch(() => nullModule));
const SettingsView = lazy(() => import("../views/SettingsView.jsx").catch(() => nullModule));

function ServerDownBanner({ networkOnline }) {
  return (
    <div className="flex flex-col items-center justify-center py-24 gap-6 text-center">
      <div className="flex flex-col items-center gap-2">
        <div className="w-12 h-12 rounded-full bg-error/10 border border-error/30 flex items-center justify-center text-2xl">
          ✕
        </div>
        <h2 className="text-xl font-semibold text-fg m-0">Server Unreachable</h2>
        <p className="text-muted text-sm m-0">goUp is not responding. It may be down or restarting.</p>
      </div>
      <div className="flex flex-col gap-2 text-sm">
        <div className="flex items-center gap-2">
          <span className={`w-2 h-2 rounded-full ${networkOnline ? "bg-success" : "bg-error"}`} />
          <span className="text-muted">Network: <span className={networkOnline ? "text-success" : "text-error"}>{networkOnline ? "Online" : "Offline"}</span></span>
        </div>
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-error" />
          <span className="text-muted">Server: <span className="text-error">Unreachable</span></span>
        </div>
      </div>
      <p className="text-muted text-xs m-0">Retrying automatically…</p>
    </div>
  );
}

class ChunkErrorBoundary extends Component {
  constructor(props) {
    super(props);
    this.state = { crashed: false };
  }

  static getDerivedStateFromError() {
    return { crashed: true };
  }

  componentDidUpdate(prevProps) {
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
    return localStorage.getItem("activeTab") || "overview";
  });
  const [serverDown, setServerDown] = useState(false);
  const [networkOnline, setNetworkOnline] = useState(navigator.onLine);

  useEffect(() => {
    localStorage.setItem("activeTab", activeTab);
  }, [activeTab]);

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
    const handleKeyDown = (e) => {
      if (e.ctrlKey || e.metaKey) {
        switch (e.key) {
          case "1":
            e.preventDefault();
            setActiveTab("overview");
            break;
          case "2":
            e.preventDefault();
            setActiveTab("services");
            break;
          case "3":
            e.preventDefault();
            setActiveTab("analytics");
            break;
          case "4":
            e.preventDefault();
            setActiveTab("settings");
            break;
          default:
            break;
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  const renderView = () => {
    if (serverDown) return <ServerDownBanner networkOnline={networkOnline} />;

    switch (activeTab) {
      case "overview":
        return <OverviewView />;
      case "services":
        return <ServicesView />;
      case "analytics":
        return <AnalyticsView />;
      case "settings":
        return <SettingsView />;
      default:
        return <OverviewView />;
    }
  };

  return (
    <div className="flex flex-col min-h-screen w-full">
      <Header />
      <Navigation activeTab={activeTab} onTabChange={setActiveTab} />
      <main
        className="flex-1 max-w-350 w-full mx-auto p-4 md:p-6 xl:p-8 animate-fade-in"
        role="tabpanel"
        id={`${activeTab}-panel`}
        aria-labelledby={`${activeTab}-tab`}
      >
        <ChunkErrorBoundary resetKey={serverDown}>
          <Suspense
            fallback={
              <div className="flex items-center justify-center py-24">
                <div className="w-6 h-6 rounded-full border-2 border-primary/20 border-t-primary animate-spin" />
              </div>
            }
          >
            {renderView()}
          </Suspense>
        </ChunkErrorBoundary>
      </main>
    </div>
  );
}
