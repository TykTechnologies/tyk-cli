# Milestone 1: US-PM-01 (List) + US-PM-02 (Get)
# All scenarios beyond walking skeleton are @pending until walking skeleton passes.

Feature: List and inspect security policies
  As Ravi Patel, a platform engineer
  I want to list all policies and inspect individual policy details
  So that I can understand the access control landscape before making changes

  Background:
    Given Ravi has a configured environment "staging"
    And the Dashboard has the following APIs:
      | api_id       | name         | listen_path |
      | a1b2c3d4e5f6 | users-api    | /users/     |
      | g7h8i9j0k1l2 | orders-api   | /orders/    |
      | m3n4o5p6q7r8 | payments-api | /payments/  |
    And the Dashboard has the following policies:
      | _id       | name        | rate | per | quota_max | quota_renewal_rate | tags       |
      | gold      | Gold Plan   | 1000 | 60  | 100000    | 2592000            | gold, paid |
      | silver    | Silver Plan | 500  | 60  | 50000     | 2592000            | silver     |
      | free-tier | Free Plan   | 100  | 60  | 10000     | 86400              | free       |

  # --- US-PM-01: List Policies ---

  @pending
  Scenario: List policies in JSON format
    When Ravi runs "tyk policy list --json"
    Then stdout contains valid JSON with an array of policy objects
    And each policy object has fields "id", "name", "api_count", "tags"
    And the JSON contains 3 policy entries
    And the exit code is 0

  @pending
  Scenario: List policies with pagination
    When Ravi runs "tyk policy list --page 2"
    And no policies exist on page 2
    Then Ravi sees "No policies found on page 2." on stderr
    And the exit code is 0

  @pending
  Scenario: List policies shows API count per policy
    Given policy "gold" has 3 APIs in access_rights
    And policy "silver" has 2 APIs in access_rights
    And policy "free-tier" has 1 API in access_rights
    When Ravi runs "tyk policy list"
    Then the table row for "gold" shows API count 3
    And the table row for "silver" shows API count 2
    And the table row for "free-tier" shows API count 1
    And the exit code is 0

  @pending
  Scenario: List policies displays tags correctly
    When Ravi runs "tyk policy list"
    Then the table row for "gold" shows tags "gold, paid"
    And the table row for "free-tier" shows tags "free"
    And the exit code is 0

  # --- US-PM-02: Get Policy Details ---

  @pending
  Scenario: Get a policy in human-readable format
    When Ravi runs "tyk policy get gold"
    Then stderr displays policy summary with name "Gold Plan"
    And stderr displays "Rate Limit: 1000 requests / 1m"
    And stderr displays "Quota: 100000 / 30d"
    And stdout contains valid YAML with "kind: Policy"
    And stdout YAML field "metadata.id" equals "gold"
    And the exit code is 0

  @pending
  Scenario: Get a policy exports CLI schema with access entries resolved to names
    When Ravi runs "tyk policy get gold"
    Then stdout YAML field "spec.access" is a list
    And access entries use "name" selectors where APIs are resolvable
    And the exit code is 0

  @pending
  Scenario: Get a policy in JSON format
    When Ravi runs "tyk policy get gold --json"
    Then stdout contains valid JSON
    And the JSON field "metadata.id" equals "gold"
    And the JSON field "metadata.name" equals "Gold Plan"
    And the JSON field "spec.rateLimit.requests" equals 1000
    And the JSON field "spec.rateLimit.per" equals "1m"
    And the JSON field "spec.quota.limit" equals 100000
    And the JSON field "spec.quota.period" equals "30d"
    And the exit code is 0

  @pending
  Scenario: Get a non-existent policy returns not-found
    When Ravi runs "tyk policy get nonexistent"
    Then stderr displays "policy 'nonexistent' not found"
    And the exit code is 3

  @pending
  Scenario: Get a policy and redirect to file for version control
    When Ravi runs "tyk policy get gold" and redirects stdout to a file
    Then the file contains valid YAML with "apiVersion: tyk.tyktech/v1"
    And the file contains "kind: Policy"
    And the file is directly usable with "tyk policy apply -f"
