import { useQuery } from "@tanstack/react-query";

function getBaseUrl(): string {
  if (typeof window !== "undefined") {
    return localStorage.getItem("nucleus-server-url") || "http://localhost:8080";
  }
  return "http://localhost:8080";
}

export function useConnection() {
  const { data, isError } = useQuery({
    queryKey: ["health"],
    queryFn: () =>
      fetch(getBaseUrl() + "/health").then((r) =>
        r.ok ? r.json() : Promise.reject(new Error("Health check failed")),
      ),
    refetchInterval: 10_000,
    retry: false,
  });

  return {
    connected: !!data && !isError,
    status: data?.status as string | undefined,
  };
}
