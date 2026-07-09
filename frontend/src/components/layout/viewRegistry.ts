import { lazy, type ComponentType } from "react";

const nullModule: { default: ComponentType } = { default: () => null };

function lazyView(loader: () => Promise<{ default: ComponentType }>) {
  return lazy(() => loader().catch(() => nullModule));
}

export const VIEWS: Record<string, ComponentType> = {
  overview: lazyView(() => import("../views/OverviewView")),
  services: lazyView(() => import("../views/ServicesView")),
  analytics: lazyView(() => import("../views/AnalyticsView")),
  settings: lazyView(() => import("../views/SettingsView")),
};

export const KEY_MAP: Record<string, string> = {
  "1": "overview",
  "2": "services",
  "3": "analytics",
  "4": "settings",
};
