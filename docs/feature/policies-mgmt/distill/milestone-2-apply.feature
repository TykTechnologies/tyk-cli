# Milestone 2: US-PM-03 (Apply)
# The most complex command: upsert, selectors, duration parsing, validation errors.

Feature: Apply security policy from file
  As Ravi Patel, a platform engineer
  I want to apply a security policy from a YAML file to the Dashboard
  So that I can manage access control as code in my GitOps workflow

  Background:
    Given Ravi has a configured environment "staging"
    And the Dashboard has the following APIs:
      | api_id       | name         | listen_path | tags          |
      | a1b2c3d4e5f6 | users-api    | /users/     | public, v1    |
      | g7h8i9j0k1l2 | orders-api   | /orders/    | internal      |
      | m3n4o5p6q7r8 | payments-api | /payments/  | public, v1    |

  # --- Selector Resolution (Happy Paths) ---

  @pending
  Scenario: Apply resolves listenPath selector to API ID
    Given Ravi has a policy file with access entry "listenPath: /orders/"
    And no policy with id "path-test" exists on the Dashboard
    When Ravi runs "tyk policy apply -f policies/path-test.yaml"
    Then the selector "listenPath: /orders/" resolves to "g7h8i9j0k1l2"
    And the Dashboard receives a create request
    And the exit code is 0

  @pending
  Scenario: Apply resolves direct ID selector
    Given Ravi has a policy file with access entry "id: a1b2c3d4e5f6"
    When Ravi runs "tyk policy apply -f policies/id-test.yaml"
    Then the selector "id: a1b2c3d4e5f6" resolves directly
    And the exit code is 0

  @pending
  Scenario: Apply resolves tags selector to multiple APIs
    Given Ravi has a policy file with access entry "tags: [public, v1]"
    When Ravi runs "tyk policy apply -f policies/tags-test.yaml"
    Then the tags selector matches "users-api" and "payments-api"
    And the Dashboard payload access_rights contains both API IDs
    And stderr displays "tags: [public, v1] -> 2 APIs matched"
    And the exit code is 0

  @pending
  Scenario: Apply with multiple access entries resolves all selectors
    Given Ravi has a policy file with access entries:
      | selector_type | value      | versions |
      | name          | users-api  | v1       |
      | listenPath    | /orders/   | v1, v2   |
      | id            | m3n4o5p6q7r8 | v1     |
    When Ravi runs "tyk policy apply -f policies/multi.yaml"
    Then all three selectors resolve successfully
    And the Dashboard payload access_rights contains 3 APIs
    And the exit code is 0

  # --- Duration Conversion ---

  @pending
  Scenario: Apply converts duration strings to seconds
    Given Ravi has a policy file with "per: 1m", "period: 30d", and "keyTTL: 24h"
    When Ravi runs "tyk policy apply -f policies/durations.yaml"
    Then the Dashboard payload has "per" equal to 60
    And the Dashboard payload has "quota_renewal_rate" equal to 2592000
    And the Dashboard payload has "key_expires_in" equal to 86400
    And the exit code is 0

  @pending
  Scenario: Apply accepts plain integer seconds
    Given Ravi has a policy file with "per: 60", "period: 2592000", and "keyTTL: 0"
    When Ravi runs "tyk policy apply -f policies/integers.yaml"
    Then the Dashboard payload has "per" equal to 60
    And the Dashboard payload has "quota_renewal_rate" equal to 2592000
    And the Dashboard payload has "key_expires_in" equal to 0
    And the exit code is 0

  @pending
  Scenario: Apply with keyTTL zero means keys never expire
    Given Ravi has a policy file with "keyTTL: 0"
    When Ravi runs "tyk policy apply -f policies/no-expiry.yaml"
    Then the Dashboard payload has "key_expires_in" equal to 0
    And the exit code is 0

  # --- Stdin Support ---

  @pending
  Scenario: Apply from stdin
    Given Ravi pipes a valid policy YAML to stdin
    When Ravi runs "tyk policy apply -f -"
    Then the policy is applied successfully
    And the exit code is 0

  # --- JSON Output ---

  @pending
  Scenario: Apply with JSON output shows structured result
    When Ravi runs "tyk policy apply -f policies/platinum.yaml --json"
    Then stdout contains valid JSON with "policy_id" and "operation"
    And the exit code is 0

  # --- Error Paths: Selector Resolution ---

  @pending
  Scenario: Apply fails when name selector matches zero APIs
    Given no API named "inventori-api" exists
    And Ravi has a policy file with access entry "name: inventori-api"
    When Ravi runs "tyk policy apply -f policies/typo.yaml"
    Then stderr displays "no API found" for the selector
    And stderr displays fuzzy suggestions including closest API names
    And stderr displays "Hint: run 'tyk api list' to see available APIs"
    And the exit code is 2

  @pending
  Scenario: Apply fails when name selector matches multiple APIs
    Given two APIs named "api-service" exist in the Dashboard
    And Ravi has a policy file with access entry "name: api-service"
    When Ravi runs "tyk policy apply -f policies/ambiguous.yaml"
    Then stderr displays "selector ambiguous" with the number of matches
    And stderr lists the matching API candidates with their IDs
    And stderr displays "use id to disambiguate"
    And the exit code is 2

  @pending
  Scenario: Apply fails when listenPath selector matches multiple APIs
    Given two APIs with listen path "/shared/" exist in the Dashboard
    And Ravi has a policy file with access entry "listenPath: /shared/"
    When Ravi runs "tyk policy apply -f policies/ambiguous-path.yaml"
    Then stderr displays "selector ambiguous" for the listenPath
    And the exit code is 2

  @pending
  Scenario: Apply fails when tags selector matches zero APIs
    Given no APIs have all tags [internal, legacy]
    And Ravi has a policy file with access entry "tags: [internal, legacy]"
    When Ravi runs "tyk policy apply -f policies/empty-tags.yaml"
    Then stderr displays "no APIs matched tags [internal, legacy]"
    And the exit code is 2

  @pending
  Scenario: Apply fails when ID selector matches no API
    Given no API with id "nonexistent-id" exists
    And Ravi has a policy file with access entry "id: nonexistent-id"
    When Ravi runs "tyk policy apply -f policies/bad-id.yaml"
    Then stderr displays "no API found for id 'nonexistent-id'"
    And the exit code is 2

  # --- Error Paths: Schema Validation ---

  @pending
  Scenario: Apply fails when metadata.id is missing
    Given Ravi has a policy file missing "metadata.id"
    When Ravi runs "tyk policy apply -f policies/no-id.yaml"
    Then stderr displays validation error for field "metadata.id"
    And stderr displays "required field missing"
    And the exit code is 2

  @pending
  Scenario: Apply fails when metadata.name is missing
    Given Ravi has a policy file missing "metadata.name"
    When Ravi runs "tyk policy apply -f policies/no-name.yaml"
    Then stderr displays validation error for field "metadata.name"
    And the exit code is 2

  @pending
  Scenario: Apply fails on invalid duration format
    Given Ravi has a policy file with "per: abc"
    When Ravi runs "tyk policy apply -f policies/bad-duration.yaml"
    Then stderr displays validation error for field "spec.rateLimit.per"
    And stderr displays "invalid duration"
    And the exit code is 2

  @pending
  Scenario: Apply fails when access entry has zero selectors
    Given Ravi has a policy file with an access entry that has no selector field
    When Ravi runs "tyk policy apply -f policies/no-selector.yaml"
    Then stderr displays "exactly one of id, name, listenPath, or tags must be set"
    And the exit code is 2

  @pending
  Scenario: Apply fails when access entry has multiple selectors
    Given Ravi has a policy file with an access entry that has both "name" and "id"
    When Ravi runs "tyk policy apply -f policies/multi-selector.yaml"
    Then stderr displays "exactly one of id, name, listenPath, or tags must be set"
    And the exit code is 2

  @pending
  Scenario: Apply fails when spec.access is empty
    Given Ravi has a policy file with empty access list
    When Ravi runs "tyk policy apply -f policies/empty-access.yaml"
    Then stderr displays validation error for "spec.access"
    And stderr displays "at least 1 access entry required"
    And the exit code is 2

  @pending
  Scenario: Apply collects multiple validation errors
    Given Ravi has a policy file missing "metadata.id" and with invalid duration "abc"
    When Ravi runs "tyk policy apply -f policies/multi-error.yaml"
    Then stderr lists all validation errors with field paths
    And the error count is greater than 1
    And the exit code is 2

  # --- Error Paths: File Handling ---

  @pending
  Scenario: Apply fails when file does not exist
    When Ravi runs "tyk policy apply -f policies/nonexistent.yaml"
    Then stderr displays "file not found"
    And the exit code is 2

  @pending
  Scenario: Apply fails when file is not valid YAML
    Given Ravi has a file "policies/bad.yaml" with invalid YAML content
    When Ravi runs "tyk policy apply -f policies/bad.yaml"
    Then stderr displays a YAML parse error
    And the exit code is 2

  @pending
  Scenario: Apply fails when no file argument is provided
    When Ravi runs "tyk policy apply"
    Then stderr displays an error about missing file argument
    And the exit code is 2
