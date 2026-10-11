import { QueryClientProvider } from "@tanstack/react-query";
import AppLayout from "./layout/AppLayout";
import { queryClient } from "../queryClient";

export default function Root() {
  return (
    <QueryClientProvider client={queryClient}>
      <AppLayout />
    </QueryClientProvider>
  );
}
