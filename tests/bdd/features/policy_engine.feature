Feature: Policy Engine
  As a trust and safety administrator
  I want to define moderation policies
  So that content enforcement is configurable and consistent

  Background:
    Given I am authenticated as an admin user

  Scenario: Create a new policy
    When I create a policy "Test Policy" with thresholds
      | category       | threshold | action |
      | toxicity       | 0.8       | block  |
      | hate           | 0.7       | block  |
      | harassment     | 0.75      | warn   |
    Then the policy should be created with status "draft"
    And the policy version should be 1

  Scenario: Publish a policy
    Given a draft policy "Test Policy" exists
    When I publish the policy
    Then the policy status should be "published"
    And the policy should have an effective date

  Scenario: Policy versioning
    Given a published policy "Standard Guidelines" version 1
    When I create a new version with updated thresholds
    Then a new version 2 should be created
    And version 1 should remain unchanged
    And historical decisions should reference version 1

  Scenario: Policy evaluation with region context
    Given a published policy scoped to region "EU"
    And a published policy scoped to region "US"
    When text is moderated with context region "EU"
    Then the EU policy should be applied

  Scenario: Unauthorized policy modification is rejected
    Given I am authenticated as a moderator user
    When I attempt to create a policy
    Then the request should be rejected with 403 Forbidden

  Scenario: Policy threshold validation
    When I attempt to create a policy with invalid thresholds
      | category   | threshold | action |
      | toxicity   | 1.5       | block  |
      | hate       | -0.2      | block  |
    Then the request should be rejected with validation errors
    And the error should indicate "thresholds must be between 0 and 1"

  Scenario: Policy action validation
    When I attempt to create a policy with invalid actions
      | category   | threshold | action        |
      | toxicity   | 0.8       | invalid_action |
    Then the request should be rejected with validation errors
    And the error should indicate "action must be allow, warn, block, or escalate"

  Scenario: Draft policy cannot be used for moderation
    Given a draft policy "Draft Policy" exists
    When text is submitted for moderation
    Then the draft policy should not be applied
    And a published policy should be used instead

  Scenario: Archive policy
    Given a published policy "Old Policy" exists
    When I archive the policy
    Then the policy status should be "archived"
    And the policy should no longer be used for new moderation decisions
    And historical evidence should still reference the archived policy
