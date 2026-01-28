import apiClient from "./client";
import type { Policy, CreatePolicyRequest, UpdatePolicyRequest } from "../types";

// Map backend policy shape to frontend Policy type
function mapPolicy(raw: any): Policy {
  return {
    id: raw.id,
    name: raw.name,
    description: raw.description,
    version: raw.version,
    status: raw.status,
    category_thresholds: raw.category_thresholds || raw.thresholds || {},
    category_actions: raw.category_actions || raw.actions || {},
    effective_date: raw.effective_date,
    created_by: raw.created_by,
    created_at: raw.created_at,
    updated_at: raw.updated_at || raw.created_at,
    published_at: raw.published_at,
  };
}

export async function listPolicies(): Promise<Policy[]> {
  const response = await apiClient.get("/policies");
  return (response.data || []).map(mapPolicy);
}

export async function getPolicy(id: string): Promise<Policy> {
  const response = await apiClient.get(`/policies/${id}`);
  return mapPolicy(response.data);
}

export async function createPolicy(policy: CreatePolicyRequest): Promise<Policy> {
  const response = await apiClient.post<Policy>("/policies", policy);
  return response.data;
}

export async function updatePolicy(id: string, policy: UpdatePolicyRequest): Promise<Policy> {
  const response = await apiClient.put<Policy>(`/policies/${id}`, policy);
  return response.data;
}

export async function publishPolicy(id: string): Promise<Policy> {
  const response = await apiClient.post<Policy>(`/policies/${id}/publish`);
  return response.data;
}

export async function archivePolicy(id: string): Promise<Policy> {
  const response = await apiClient.post<Policy>(`/policies/${id}/archive`);
  return response.data;
}
