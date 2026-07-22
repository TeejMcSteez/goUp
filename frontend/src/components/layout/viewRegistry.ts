import { lazy, type ComponentType } from "react";

const nullModule: { default: ComponentType } = { default: () => null };
const RELOAD_KEY = "goup:chunk-reload";

function lazyView(loader: () => Promise<{ default: ComponentType }>) {
  return lazy(() =>
    loader()
      .then((mod) => {
        sessionStorage.removeItem(RELOAD_KEY);
        return mod;
      })
      .catch((err) => {
        // Stale chunk hash after a rebuild (old tab, new dist): reload once
        // to pick up the fresh index.html instead of rendering a blank view.
        if (!sessionStorage.getItem(RELOAD_KEY)) {
          sessionStorage.setItem(RELOAD_KEY, "1");
          window.location.reload();
          return new Promise<{ default: ComponentType }>(() => {});
        }
        console.error(err);
        return nullModule;
      }),
  );
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
