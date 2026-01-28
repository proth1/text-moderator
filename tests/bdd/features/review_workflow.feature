Feature: Human Review Workflow
  As a content moderator
  I want to review flagged content
  So that I can override incorrect automated decisions

  Background:
    Given I am authenticated as a moderator user
    And there are pending moderation decisions in the queue

  Scenario: View moderation queue
    When I request the moderation queue
    Then I should see a list of pending decisions
    And each decision should show category, confidence, and status

  Scenario: Approve a block decision
    Given a moderation decision with action "block"
    When I approve the decision with rationale "Confirmed violation"
    Then the review action should be recorded
    And an evidence record with control "GOV-002" should be created

  Scenario: Override a block decision
    Given a moderation decision with action "block"
    When I override the decision to "allow" with rationale "False positive - quoted text"
    Then the override should be recorded as structured data
    And an evidence record should capture the human override

  Scenario: Escalate a decision
    Given a moderation decision requiring further review
    When I escalate the decision
    Then the decision should require admin approval
    And the escalation should be logged

  Scenario: Review with edited content
    Given a moderation decision with action "warn"
    When I edit the content and approve with action "edit"
    Then the edited content should be stored
    And an evidence record should capture the content modification

  Scenario: Multiple reviews on same decision
    Given a moderation decision with action "block"
    And the decision was previously reviewed and rejected
    When I approve the decision with rationale "Reconsidered - clear violation"
    Then both review actions should be recorded
    And the most recent review should take precedence

  Scenario: Unauthorized review is rejected
    Given I am authenticated as a viewer user
    When I attempt to submit a review action
    Then the request should be rejected with 403 Forbidden

  Scenario: Review queue filtering
    Given multiple moderation decisions exist
      | action   | confidence | status  |
      | block    | 0.95       | pending |
      | warn     | 0.78       | pending |
      | escalate | 0.82       | pending |
      | block    | 0.88       | reviewed |
    When I request the moderation queue with filter "action=block&status=pending"
    Then I should see only pending block decisions
    And the results should be sorted by created_at descending

  Scenario: Rationale is required for override
    Given a moderation decision with action "block"
    When I attempt to override the decision without a rationale
    Then the request should be rejected with validation errors
    And the error should indicate "rationale is required for overrides"
