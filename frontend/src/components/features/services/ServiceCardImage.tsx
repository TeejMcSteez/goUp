import { ReactNode } from "react";
import fallback from "../../../../static/goup.webp";

interface ServiceCardImageProps {
  faviconUrl: string
}

export default function ServiceCardImage(props: ServiceCardImageProps): ReactNode {
  const mediaQuery = matchMedia("(width <= 640px)");
  if (mediaQuery.matches) {
    // mobile return blank to not crowd name
    return <></>
  }
  return <img
                src={props.faviconUrl}
                alt="favicon"
                className="hidden md:flex w-4 h-4 rounded-sm shrink-0 grayscale"
                onError={(e) => {
                  e.currentTarget.src = fallback;
                }}
              />
}
