import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import AppLayout from "./layout/AppLayout";

const queryClient = new QueryClient();

export default function Root() {
  return (
    <QueryClientProvider client={queryClient}>
      <AppLayout />
    </QueryClientProvider>
  );
}
