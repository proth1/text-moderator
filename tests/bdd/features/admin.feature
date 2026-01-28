Feature: Admin and Governance
  As an administrator
  I want to manage policies and access controls
  So that the platform operates within governance requirements

  Scenario: Admin can access policy management
    Given I am authenticated as an admin user
    When I access the policy management page
    Then I should see all policies listed
    And I should be able to create new policies

  Scenario: RBAC enforcement
    Given users with different roles exist
    When each user attempts to access policy management
    Then only admin users should have write access
    And moderator users should have read-only access
    And viewer users should be denied access

  Scenario: Audit log is searchable
    Given multiple evidence records exist
    When I search the audit log by control_id "MOD-001"
    Then I should see all matching records
    And the results should be sortable and filterable

  Scenario: Admin can view user activity
    Given multiple users have performed actions
    When I request the user activity report
    Then I should see a list of actions by user
    And each action should include timestamp and action type

  Scenario: Admin can manage user roles
    Given a user exists with role "viewer"
    When I update the user role to "moderator"
    Then the user role should be changed
    And the user should have moderator permissions

  Scenario: Admin cannot delete users with active reviews
    Given a moderator user has submitted review actions
    When I attempt to delete the moderator user
    Then the deletion should be rejected
    And the error should indicate "user has active review history"

  Scenario: System health check
    Given the moderation service is running
    When I request the system health status
    Then I should see the status of all services
      | service    | status |
      | database   | healthy |
      | redis      | healthy |
      | api        | healthy |
    And the response should include version information

  Scenario: Evidence export for compliance
    Given evidence records exist for the date range "2026-01-01" to "2026-01-31"
    When I export evidence records for compliance
    Then the export should include all records in the date range
    And the export format should be CSV
    And the export should include digital signature for integrity

  Scenario: Policy impact analysis
    Given a published policy "Standard Guidelines" is active
    And moderation decisions have been made using the policy
    When I request policy impact analysis
    Then I should see statistics for the policy
      | metric                  | value |
      | total_decisions         | 150   |
      | blocked_count           | 45    |
      | warned_count            | 30    |
      | allowed_count           | 75    |
      | average_confidence      | 0.87  |
    And the analysis should include category breakdown

  Scenario: Admin dashboard metrics
    Given the moderation service has processed content
    When I access the admin dashboard
    Then I should see real-time metrics
      | metric                     | displayed |
      | moderation_requests_today  | yes       |
      | average_response_time      | yes       |
      | cache_hit_rate             | yes       |
      | pending_reviews_count      | yes       |
      | model_accuracy_rate        | yes       |
