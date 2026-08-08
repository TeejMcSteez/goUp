import { useState } from "react";
import fireManual from "../hooks/fireManual";

export default function RefreshButton() {
  const [error, setError] = useState(null)
  const fire = () => {
    fireManual().catch((e) => {
      setError(e)
    })
    setTimeout(() => {
      window.location.reload()
    }, 750)
  }

  const link = (
    <a
      onClick={fire}
      className="cursor-pointer no-underline text-center text-lg text-fg px-6 py-2 rounded-lg border border-border bg-surface transition-colors duration-200 hover:bg-hover hover:text-primary"
    >
      Refresh
    </a>
  )

  if (error) {
    return (
      <div className="flex flex-col items-center">
        {link}
        <p className="m-0 text-error text-xs">Refresh Failed</p>
      </div>
    )
  }

  return <div className="flex flex-col items-center">{link}</div>
}
