import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api"
import type { components } from "@/lib/api/v1"

type User = components["schemas"]["Tenant"]

export function useUser() {
  return useQuery({
    queryKey: ["user", "me"],
    queryFn: async () => {
      const { data, error } = await api.GET("/me")
      if (error) throw error
      if (!data) throw new Error("No user returned")
      return data as User
    },
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
    retry: false, // Don't retry on 401s
  })
}
