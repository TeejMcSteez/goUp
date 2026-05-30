interface Tab {
  id: string;
  label: string;
  key: string;
}

interface NavigationProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
}

const tabs: Tab[] = [
  { id: "overview", label: "Overview", key: "1" },
  { id: "services", label: "Services", key: "2" },
  { id: "analytics", label: "Analytics", key: "3" },
  { id: "settings", label: "Settings", key: "4" },
];

export default function Navigation({
  activeTab,
  onTabChange,
}: NavigationProps) {
  const activeIndex = tabs.findIndex((t) => t.id === activeTab);
  const thumbPercent = (activeIndex / tabs.length) * 100;
  const thumbWidth = 100 / tabs.length;

  return (
    <nav
      className="sticky top-0 z-100 bg-app-bg border-b border-border px-4"
      role="tablist"
      aria-label="Main navigation"
    >
      <div className="flex gap-1 max-w-350 mx-auto overflow-x-auto">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            role="tab"
            aria-selected={activeTab === tab.id}
            aria-controls={`${tab.id}-panel`}
            id={`${tab.id}-tab`}
            className={`relative px-3 py-3 bg-transparent border-none text-xs font-medium cursor-pointer whitespace-nowrap transition-colors duration-200 hover:text-fg focus-visible:outline focus-visible:outline-primary focus-visible:-outline-offset-2 sm:px-6 sm:py-4 sm:text-sm ${
              activeTab === tab.id ? "text-primary tab-active" : "text-muted"
            }`}
            onClick={() => onTabChange(tab.id)}
            tabIndex={activeTab === tab.id ? 0 : -1}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Position indicator track */}
      <div className="relative h-0.5 bg-border max-w-350 mx-auto sm:hidden">
        <div
          className="absolute top-0 h-full bg-primary rounded-full transition-[left,width] duration-200"
          style={{
            left: `${thumbPercent}%`,
            width: `${thumbWidth}%`,
          }}
        />
      </div>
    </nav>
  );
}
