import { useState, useCallback, Suspense } from "react";
import Header from "../Header";
import Navigation from "./Navigation";
import ChunkErrorBoundary from "./ChunkErrorBoundary";
import ServerDownBanner from "./ServerDownBanner";
import WindowScrollButton from "./WindowScrollButton";
import { VIEWS, KEY_MAP } from "./viewRegistry";
import { useServerHealth } from "../../hooks/useServerHealth";
import { useTabKeyboardNav } from "../../hooks/useTabKeyboardNav";

const SPINNER = (
  <div className="flex items-center justify-center py-24">
    <div className="w-6 h-6 rounded-full border-2 border-primary/20 border-t-primary animate-spin" />
  </div>
);

export default function AppLayout() {
  const [activeTab, setActiveTab] = useState(() => {
    return localStorage.getItem("activeTab") ?? "overview";
  });
  const { serverDown, networkOnline } = useServerHealth();

  const handleTabChange = useCallback((tab: string) => {
    localStorage.setItem("activeTab", tab);
    setActiveTab(tab);
  }, []);

  useTabKeyboardNav(activeTab, handleTabChange, KEY_MAP);

  const ActiveView = VIEWS[activeTab] ?? VIEWS.overview;

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
