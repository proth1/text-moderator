Feature: Text Moderation
  As a platform operator
  I want text to be automatically moderated
  So that harmful content is detected and actioned

  Background:
    Given the moderation service is running
    And a published policy "Standard Community Guidelines" exists

  Scenario: Safe text is allowed
    Given a text submission "Hello, this is a friendly message!"
    When the text is submitted for moderation
    Then the moderation decision should be "allow"
    And all category scores should be below 0.5
    And an evidence record with control "MOD-001" should be created

  Scenario: Toxic text is blocked
    Given a text submission with high toxicity scores
    When the text is submitted for moderation
    Then the moderation decision should be "block"
    And the toxicity score should be above 0.8
    And an evidence record with control "MOD-001" should be created

  Scenario: Borderline text triggers warning
    Given a text submission with moderate harassment scores
    When the text is submitted for moderation
    Then the moderation decision should be "warn"
    And a human-readable explanation should be provided

  Scenario: API timeout is handled gracefully
    Given the HuggingFace API is unavailable
    When the text is submitted for moderation
    Then the response should indicate a service error
    And the error should be logged with correlation ID

  Scenario: Cached results are returned for repeated text
    Given a text "Hello world" was previously moderated
    When the same text is submitted again
    Then the cached result should be returned
    And the response time should be under 50ms

  Scenario: High confidence block decision
    Given a text submission with extreme hate speech
    When the text is submitted for moderation
    Then the moderation decision should be "block"
    And the confidence score should be above 0.9
    And the category "hate" score should be above 0.9

  Scenario: Multiple policy categories triggered
    Given a text submission that is both toxic and profane
    When the text is submitted for moderation
    Then the moderation decision should be "block"
    And multiple category thresholds should be exceeded
    And the strictest action should be applied
