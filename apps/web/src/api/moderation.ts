import apiClient from "./client";
import type { ModerationRequest, ModerationResponse } from "../types";

export async function moderateText(
  text: string,
  context?: Record<string, any>
): Promise<ModerationResponse> {
  const request: ModerationRequest = {
    text,
    context,
  };

  const response = await apiClient.post<ModerationResponse>("/moderate", request);
  return response.data;
}

export async function getModerationHistory(limit = 50) {
  const response = await apiClient.get("/moderate/history", {
    params: { limit },
  });
  return response.data;
}
