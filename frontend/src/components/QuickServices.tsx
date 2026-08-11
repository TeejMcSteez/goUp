import useQuickServices from "../hooks/useQuickServices";
import RefreshButton from "./RefreshButton";

export default function QuickServices() {
  const { data, error } = useQuickServices();
  const qs = data ?? [];

  function navToServices() {
    window.dispatchEvent(
      new CustomEvent("goup:navigate", { detail: "services" }),
    );
  }
  function navToOverview() {
    window.dispatchEvent(
      new CustomEvent("goup:navigate", { detail: "overview" }),
    );
  }
  const contentDivStyles = "flex flex-row items-center justify-center"
  const content =
    qs.length === 0 ? (
      <div className={contentDivStyles}>
        <a
          href="#cards"
          onClick={() => navToOverview()}
          className="no-underline text-center text-lg text-fg px-6 py-2 rounded-lg transition-colors duration-200 hover:bg-hover"
        >
          All Systems Operational ✅
        </a>
        <RefreshButton />
      </div>
    ) : (
        <div className={contentDivStyles}>
          <a
            href="#"
            onClick={() => navToServices()}
            className="no-underline text-center text-lg text-fg px-6 py-2 rounded-lg transition-colors duration-200 hover:bg-hover"
          >
            {qs.length} Errors Detected ❌
          </a>
          <RefreshButton />
      </div>
    );

  return (
    <div className="flex flex-row items-center justify-center text-center p-4 border-b border-border bg-surface">
      {error ? (
        <span className="text-error text-xs flex flex-row justify-center items-center">
          Status unavailable: {error}
        </span>
      ) : (
        content
      )}
    </div>
  );
}
