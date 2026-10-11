import { QueryClient } from "@tanstack/react-query";
import { createLiveConnection } from "./hooks/liveConnection";

// App-wide singletons, kept out of Root.tsx so component files only export
// components.
export const queryClient = new QueryClient();
export const liveConnection = createLiveConnection(queryClient);
