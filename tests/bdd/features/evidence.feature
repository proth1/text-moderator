Feature: Evidence and Auditability
  As a compliance officer
  I want all moderation decisions to generate evidence
  So that actions are traceable and auditable

  Scenario: Automated decision generates evidence
    Given a text submission is moderated
    When the automated decision is made
    Then an evidence record should be created
    And the record should include control_id, policy_id, and decision_id
    And the record should be marked as immutable

  Scenario: Evidence cannot be modified
    Given an existing evidence record
    When an attempt is made to modify the record
    Then the modification should be rejected
    And the original record should remain unchanged

  Scenario: Evidence export matches stored data
    Given multiple evidence records exist
    When I export evidence records
    Then the exported data should exactly match stored records
    And the export should include all required audit fields

  Scenario: Evidence references correct policy version
    Given a policy "Standard Guidelines" version 2 is active
    When a moderation decision is made
    Then the evidence should reference policy version 2

  Scenario: Evidence cannot be deleted
    Given an existing evidence record
    When an attempt is made to delete the record
    Then the deletion should be rejected
    And the record should remain in the database

  Scenario: Human review generates additional evidence
    Given a moderation decision exists
    When a human reviewer overrides the decision
    Then a new evidence record with control "GOV-002" should be created
    And the evidence should link to both the decision and the review action

  Scenario: Evidence includes model version
    Given a text submission is moderated
    When the automated decision is made
    Then the evidence record should include model_name and model_version
    And the model version should match the active inference model

  Scenario: Evidence search by control ID
    Given evidence records exist for multiple controls
      | control_id | count |
      | MOD-001    | 15    |
      | GOV-002    | 8     |
      | AUD-001    | 5     |
    When I search evidence by control_id "MOD-001"
    Then I should see exactly 15 evidence records
    And all records should have control_id "MOD-001"

  Scenario: Evidence timestamp immutability
    Given an evidence record was created at "2026-01-15T10:00:00Z"
    When I retrieve the evidence record
    Then the created_at timestamp should be "2026-01-15T10:00:00Z"
    And the timestamp should not be updatable

  Scenario: Evidence chain of custody
    Given a text submission is moderated and blocked
    And the decision is reviewed and approved by a moderator
    When I retrieve the evidence chain for the submission
    Then I should see evidence for control "MOD-001" with automated action
    And I should see evidence for control "GOV-002" with human review
    And both evidence records should reference the same submission hash
