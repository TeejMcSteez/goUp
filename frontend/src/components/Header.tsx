import QuickServices from "./QuickServices";

export default function Header() {
  return (
    <header className="sticky top-0 z-200 bg-app-bg border-b border-border">
      {/*<div className="text-center pt-4 px-4 pb-4 bg-app-bg">
        <h1 className="m-0 text-[2.5rem] text-primary">GoUp</h1>
        <h2 className="mt-2 mb-0 text-base text-muted font-normal">
          Service Monitoring Dashboard
        </h2>
      </div>*/}
      <QuickServices />
    </header>
  );
}
