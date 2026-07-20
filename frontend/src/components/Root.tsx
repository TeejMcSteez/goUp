import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { PollRateProvider } from "../context/PollRateContext";
import AppLayout from "./layout/AppLayout";

const queryClient = new QueryClient();

export default function Root() {
  return (
    <QueryClientProvider client={queryClient}>
      <PollRateProvider>
        <AppLayout />
      </PollRateProvider>
    </QueryClientProvider>
  );
}
