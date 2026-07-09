import { useEffect } from "react";

export function useTabKeyboardNav(
  activeTab: string,
  onTabChange: (tab: string) => void,
  keyMap: Record<string, string>,
) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key in keyMap) {
        e.preventDefault();
        onTabChange(keyMap[e.key]);
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
        const views = Object.values(keyMap);
        const current = views.indexOf(activeTab);
        if (e.key === "ArrowLeft") {
          onTabChange(views[(current - 1 + views.length) % views.length]);
        } else {
          onTabChange(views[(current + 1) % views.length]);
        }
      }
    };

    const handleNavigate = (e: Event) =>
      onTabChange((e as CustomEvent<string>).detail);

    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("goup:navigate", handleNavigate);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("goup:navigate", handleNavigate);
    };
  }, [activeTab, onTabChange, keyMap]);
}
