import { createContext, use, useState, type ReactNode } from "react";

const DEFAULT_POLL_RATE = 5000;
const STORAGE_KEY = "goUp_poll_rate";

function loadStoredPollRate(): number {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored === null) return DEFAULT_POLL_RATE;
  const val = Number(stored);
  return Number.isFinite(val) && val > 0 ? val : DEFAULT_POLL_RATE;
}

interface PollRateContextValue {
  pollRate: number;
  setPollRate: (n: number) => void;
}

const PollRateContext = createContext<PollRateContextValue>({
  pollRate: DEFAULT_POLL_RATE,
  setPollRate: () => {},
});

export function PollRateProvider({ children }: { children: ReactNode }) {
  const [pollRate, setPollRateState] = useState(loadStoredPollRate);

  function setPollRate(n: number) {
    localStorage.setItem(STORAGE_KEY, String(n));
    setPollRateState(n);
  }

  return (
    <PollRateContext value={{ pollRate, setPollRate }}>
      {children}
    </PollRateContext>
  );
}

export function usePollRate(): PollRateContextValue {
  return use(PollRateContext);
}
