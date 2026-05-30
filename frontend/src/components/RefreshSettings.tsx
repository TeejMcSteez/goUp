import { useRef } from "react";
import { usePollRate } from "../context/PollRateContext";

export default function RefreshSettingsView() {
  const { pollRate, setPollRate } = usePollRate();
  const inputRef = useRef<HTMLInputElement>(null);

  function updatePollRate() {
    if (inputRef.current === null) return;
    const val = Number(inputRef.current.value);
    if (!Number.isFinite(val) || val <= 0) return;
    setPollRate(val);
  }

  return (
    <div className="pt-3 px-1 m-1">
      <h4>Change Refresh Rate</h4>
      <p className="text-xs text-shadow-2xs">rate is in milliseconds</p>
      <input
        className="px-3 pt-1"
        name="updateRefreshInput"
        id="updateRefreshInput"
        ref={inputRef}
        placeholder={pollRate.toString()}
      />
      <button onClick={updatePollRate}>Change</button>
    </div>
  );
}
