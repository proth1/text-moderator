import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { moderateText, getModerationHistory } from "../api/moderation";

export function useModerateText() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ text, context }: { text: string; context?: Record<string, any> }) =>
      moderateText(text, context),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["moderation-history"] });
    },
  });
}

export function useModerationHistory(limit = 50) {
  return useQuery({
    queryKey: ["moderation-history", limit],
    queryFn: () => getModerationHistory(limit),
  });
}
