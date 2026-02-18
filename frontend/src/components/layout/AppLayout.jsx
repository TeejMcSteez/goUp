import { useState, useEffect } from "react";
import Header from "../Header.jsx";
import Navigation from "./Navigation.jsx";
import OverviewView from "../views/OverviewView.jsx";
import ServicesView from "../views/ServicesView.jsx";
import AnalyticsView from "../views/AnalyticsView.jsx";
import SettingsView from "../views/SettingsView.jsx";

export default function AppLayout() {
  const [activeTab, setActiveTab] = useState(() => {
    return localStorage.getItem("activeTab") || "overview";
  });

  useEffect(() => {
    localStorage.setItem("activeTab", activeTab);
  }, [activeTab]);

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
    <div className="app-layout">
      <Header />
      <Navigation activeTab={activeTab} onTabChange={setActiveTab} />
      <main
        className="view-container"
        role="tabpanel"
        id={`${activeTab}-panel`}
        aria-labelledby={`${activeTab}-tab`}
      >
        {renderView()}
      </main>
    </div>
  );
}
