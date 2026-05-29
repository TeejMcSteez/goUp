import { useState, useEffect } from "react";
import type { ServiceCardProps } from "../types";
import fallback from "../../static/goup.png";

export default function useServiceIcon(s: ServiceCardProps): string {
  const [iconUrl, setIconUrl] = useState<string>(fallback);

  useEffect(() => {
    const controller = new AbortController();
    const base = s.service.url.endsWith("/")
      ? s.service.url
      : s.service.url + "/";
    const faviconUrl = base + "favicon.ico";

    fetch(faviconUrl, { signal: controller.signal })
      .then((res) => setIconUrl(res.ok ? faviconUrl : fallback))
      .catch((err: Error) => {
        if (err.name !== "AbortError") setIconUrl(fallback);
      });

    return () => controller.abort();
  }, [s.service.url]);

  return iconUrl;
}
