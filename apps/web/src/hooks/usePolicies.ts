import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  listPolicies,
  getPolicy,
  createPolicy,
  updatePolicy,
  publishPolicy,
  archivePolicy,
} from "../api/policies";
import type { CreatePolicyRequest, UpdatePolicyRequest } from "../types";

export function usePolicies() {
  return useQuery({
    queryKey: ["policies"],
    queryFn: listPolicies,
  });
}

export function usePolicy(id: string) {
  return useQuery({
    queryKey: ["policy", id],
    queryFn: () => getPolicy(id),
    enabled: !!id,
  });
}

export function useCreatePolicy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (policy: CreatePolicyRequest) => createPolicy(policy),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["policies"] });
    },
  });
}

export function useUpdatePolicy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, policy }: { id: string; policy: UpdatePolicyRequest }) =>
      updatePolicy(id, policy),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["policies"] });
      queryClient.invalidateQueries({ queryKey: ["policy", variables.id] });
    },
  });
}

export function usePublishPolicy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => publishPolicy(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ["policies"] });
      queryClient.invalidateQueries({ queryKey: ["policy", id] });
    },
  });
}

export function useArchivePolicy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => archivePolicy(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ["policies"] });
      queryClient.invalidateQueries({ queryKey: ["policy", id] });
    },
  });
}
