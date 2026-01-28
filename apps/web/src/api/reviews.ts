import apiClient from "./client";
import type { ReviewItem, ReviewDetail, ReviewFilters, ReviewActionType } from "../types";

export async function listReviews(filters?: ReviewFilters): Promise<ReviewItem[]> {
  const response = await apiClient.get<ReviewItem[]>("/reviews", {
    params: filters,
  });
  return response.data;
}

export async function getReview(id: string): Promise<ReviewDetail> {
  const response = await apiClient.get<ReviewDetail>(`/reviews/${id}`);
  return response.data;
}

export async function submitAction(
  id: string,
  action: ReviewActionType,
  rationale?: string
): Promise<void> {
  await apiClient.post(`/reviews/${id}/action`, {
    action_type: action,
    rationale,
  });
}
