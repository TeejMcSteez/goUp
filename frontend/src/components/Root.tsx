import { PollRateProvider } from "../context/PollRateContext";
import AppLayout from "./layout/AppLayout";

export default function Root() {
  return (
    <PollRateProvider>
      <AppLayout />
    </PollRateProvider>
  );
}
