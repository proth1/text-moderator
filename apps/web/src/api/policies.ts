import apiClient from "./client";
import type { Policy, CreatePolicyRequest, UpdatePolicyRequest } from "../types";

export async function listPolicies(): Promise<Policy[]> {
  const response = await apiClient.get<Policy[]>("/policies");
  return response.data;
}

export async function getPolicy(id: string): Promise<Policy> {
  const response = await apiClient.get<Policy>(`/policies/${id}`);
  return response.data;
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
