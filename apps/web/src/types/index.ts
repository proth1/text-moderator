// Enums (using const enums for erasableSyntaxOnly compatibility)
export const PolicyAction = {
  ALLOW: "allow",
  WARN: "warn",
  BLOCK: "block",
  REVIEW: "review",
} as const;

export type PolicyAction = typeof PolicyAction[keyof typeof PolicyAction];

export const PolicyStatus = {
  DRAFT: "draft",
  PUBLISHED: "published",
  ARCHIVED: "archived",
} as const;

export type PolicyStatus = typeof PolicyStatus[keyof typeof PolicyStatus];

export const ReviewActionType = {
  APPROVE: "approve",
  OVERRIDE_ALLOW: "override_allow",
  OVERRIDE_BLOCK: "override_block",
  ESCALATE: "escalate",
} as const;

export type ReviewActionType = typeof ReviewActionType[keyof typeof ReviewActionType];

export const UserRole = {
  ADMIN: "admin",
  MODERATOR: "moderator",
  VIEWER: "viewer",
} as const;

export type UserRole = typeof UserRole[keyof typeof UserRole];

// Core types
export interface CategoryScores {
  hate_speech?: number;
  harassment?: number;
  violence?: number;
  sexual_content?: number;
  spam?: number;
  misinformation?: number;
  pii?: number;
  [key: string]: number | undefined;
}

export interface TextSubmission {
  id: string;
  text: string;
  context?: Record<string, any>;
  user_id?: string;
  metadata?: Record<string, any>;
  created_at: string;
}

export interface ModerationDecision {
  id: string;
  submission_id: string;
  category_scores: CategoryScores;
  suggested_action: PolicyAction;
  confidence_score: number;
  policy_id?: string;
  policy_version?: number;
  reasons: string[];
  metadata?: Record<string, any>;
  created_at: string;
}

export interface Policy {
  id: string;
  name: string;
  description?: string;
  version: number;
  status: PolicyStatus;
  category_thresholds: CategoryScores;
  category_actions: Record<string, PolicyAction>;
  effective_date?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
  published_at?: string;
}

export interface ReviewAction {
  id: string;
  decision_id: string;
  action_type: ReviewActionType;
  reviewer_id: string;
  rationale?: string;
  metadata?: Record<string, any>;
  created_at: string;
}

export interface EvidenceRecord {
  id: string;
  control_id: string;
  policy_id?: string;
  policy_version?: number;
  decision_id?: string;
  review_id?: string;
  model_name?: string;
  model_version?: string;
  category_scores?: CategoryScores;
  automated_action?: string;
  human_override?: string;
  submission_hash?: string;
  immutable: boolean;
  created_at: string;
}

export interface User {
  id: string;
  email: string;
  role: UserRole;
  api_key?: string;
  created_at: string;
}

// API Request/Response types
export interface ModerationRequest {
  content: string;
  context?: Record<string, any>;
  user_id?: string;
}

export interface ModerationResponse {
  submission_id: string;
  decision_id: string;
  action: PolicyAction;
  category_scores: CategoryScores;
  confidence: number;
  reasons: string[];
  requires_review: boolean;
}

export interface ReviewItem {
  id: string;
  decision_id: string;
  submission_id: string;
  text: string;
  category_scores: CategoryScores;
  suggested_action: PolicyAction;
  confidence: number;
  status: "pending" | "reviewed";
  created_at: string;
}

export interface ReviewDetail extends ReviewItem {
  policy_id?: string;
  policy_name?: string;
  reasons: string[];
  actions: ReviewAction[];
}

export interface CreatePolicyRequest {
  name: string;
  description?: string;
  category_thresholds: CategoryScores;
  category_actions: Record<string, PolicyAction>;
  effective_date?: string;
}

export interface UpdatePolicyRequest extends Partial<CreatePolicyRequest> {}

export interface ReviewFilters {
  status?: "pending" | "reviewed";
  category?: string;
  min_confidence?: number;
  max_confidence?: number;
  from_date?: string;
  to_date?: string;
}

export interface EvidenceFilters {
  control_id?: string;
}

export interface DashboardStats {
  total_moderations: number;
  blocked_percentage: number;
  pending_reviews: number;
  active_policies: number;
  moderations_today: number;
  reviews_today: number;
}
