import apiClient from "./client";
import type { EvidenceRecord, EvidenceFilters } from "../types";

export async function listEvidence(filters?: EvidenceFilters): Promise<EvidenceRecord[]> {
  const response = await apiClient.get<EvidenceRecord[]>("/evidence", {
    params: filters,
  });
  return response.data;
}

export async function exportEvidence(filters?: EvidenceFilters): Promise<Blob> {
  const response = await apiClient.get("/evidence/export", {
    params: filters,
    responseType: "blob",
  });
  return response.data;
}
