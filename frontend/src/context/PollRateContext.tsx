import { createContext, use, useState, type ReactNode } from "react";

const DEFAULT_POLL_RATE = 5000;

interface PollRateContextValue {
  pollRate: number;
  setPollRate: (n: number) => void;
}

const PollRateContext = createContext<PollRateContextValue>({
  pollRate: DEFAULT_POLL_RATE,
  setPollRate: () => {},
});

export function PollRateProvider({ children }: { children: ReactNode }) {
  const [pollRate, setPollRate] = useState(DEFAULT_POLL_RATE);
  return (
    <PollRateContext value={{ pollRate, setPollRate }}>
      {children}
    </PollRateContext>
  );
}

export function usePollRate(): PollRateContextValue {
  return use(PollRateContext);
}
