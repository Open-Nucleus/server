import { useQuery } from '@tanstack/react-query';
import { getServerUrl } from '@/stores/auth-store';

export function useConnection() {
  const { data, isError } = useQuery({
    queryKey: ['health'],
    queryFn: async () => {
      const base = getServerUrl() || 'http://localhost:8080';
      const res = await fetch(base + '/health');
      if (!res.ok) throw new Error('Health check failed');
      return res.json();
    },
    refetchInterval: 10_000,
    retry: false,
  });

  return {
    connected: !!data && !isError,
    status: data?.status as string | undefined,
  };
}
