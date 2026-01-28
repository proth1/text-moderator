import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { listReviews, getReview, submitAction } from "../api/reviews";
import type { ReviewFilters, ReviewActionType } from "../types";

export function useReviewQueue(filters?: ReviewFilters) {
  return useQuery({
    queryKey: ["reviews", filters],
    queryFn: () => listReviews(filters),
  });
}

export function useReviewDetail(id: string) {
  return useQuery({
    queryKey: ["review", id],
    queryFn: () => getReview(id),
    enabled: !!id,
  });
}

export function useSubmitReviewAction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      action,
      rationale,
    }: {
      id: string;
      action: ReviewActionType;
      rationale?: string;
    }) => submitAction(id, action, rationale),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["reviews"] });
      queryClient.invalidateQueries({ queryKey: ["review", variables.id] });
    },
  });
}
