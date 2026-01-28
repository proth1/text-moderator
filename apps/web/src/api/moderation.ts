import apiClient from "./client";
import type { ModerationResponse } from "../types";

export async function moderateText(
  text: string,
  context?: Record<string, any>
): Promise<ModerationResponse> {
  // Backend expects 'content' field, not 'text'
  const response = await apiClient.post("/moderate", {
    content: text,
    context,
  });

  const data = response.data;
  const categoryScores = data.category_scores || {};

  // Derive confidence from max category score (backend doesn't return it directly)
  const scores = Object.values(categoryScores).filter(
    (v): v is number => typeof v === "number"
  );
  const confidence = scores.length > 0 ? Math.max(...scores) : 0;

  return {
    submission_id: data.submission_id,
    decision_id: data.decision_id,
    action: data.action,
    category_scores: categoryScores,
    confidence,
    reasons: data.reasons || [],
    requires_review: data.requires_review ?? false,
  };
}

export async function getModerationHistory(limit = 50) {
  const response = await apiClient.get("/moderate/history", {
    params: { limit },
  });
  return response.data;
}
