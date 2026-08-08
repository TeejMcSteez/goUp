interface Response {
  status: string
}

export default async function fireManual(): Promise<string> {
  const res = await fetch("/api/fire", {method: "POST"})
  if (!res.ok) {

  }
  const json: Response = await res.json()
  if (json.status != 'success') {
    return "failed to refresh"
  } else {
    return "refreshed"
  }
}
