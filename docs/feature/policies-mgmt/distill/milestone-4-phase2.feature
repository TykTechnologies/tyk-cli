# Milestone 4: US-PM-06 (Who-Uses) + US-PM-07 (Bind/Unbind)
# Phase 2 features -- depends on Phase 1 completion.

Feature: Cross-reference policies with APIs and quick-edit access rights
  As Ravi Patel, a platform engineer
  I want to check which policies reference an API before modifying it
  And quickly add or remove APIs from policies without a full apply cycle
  So that I can perform impact analysis and make urgent access changes safely

  Background:
    Given Ravi has a configured environment "staging"
    And the Dashboard has the following APIs:
      | api_id       | name         | listen_path |
      | a1b2c3d4e5f6 | users-api    | /users/     |
      | g7h8i9j0k1l2 | orders-api   | /orders/    |
      | m3n4o5p6q7r8 | payments-api | /payments/  |
    And the Dashboard has the following policies:
      | _id    | name        | access_rights_apis          |
      | gold   | Gold Plan   | a1b2c3d4e5f6, g7h8i9j0k1l2 |
      | silver | Silver Plan | a1b2c3d4e5f6                |

  # --- US-PM-06: Who-Uses ---

  @pending
  Scenario: Show policies referencing an API by name
    When Ravi runs "tyk api who-uses users-api"
    Then stdout displays a table with policies referencing "users-api"
    And stdout contains "gold" and "Gold Plan"
    And stdout contains "silver" and "Silver Plan"
    And the exit code is 0

  @pending
  Scenario: Show policies referencing an API by ID
    When Ravi runs "tyk api who-uses a1b2c3d4e5f6"
    Then stdout displays the same referencing policies as by-name lookup
    And the exit code is 0

  @pending
  Scenario: Show policies referencing an API by listenPath
    When Ravi runs "tyk api who-uses /users/"
    Then stdout displays policies referencing the API at "/users/"
    And the exit code is 0

  @pending
  Scenario: No policies reference the given API
    When Ravi runs "tyk api who-uses payments-api"
    Then stderr displays "No policies reference this API."
    And the exit code is 0

  @pending
  Scenario: Who-uses for a non-existent API
    When Ravi runs "tyk api who-uses nonexistent-api"
    Then stderr displays "No API matching 'nonexistent-api' found"
    And the exit code is 3

  @pending
  Scenario: Who-uses with JSON output
    When Ravi runs "tyk api who-uses users-api --json"
    Then stdout contains valid JSON with an array of referencing policies
    And each entry has "policy_id", "policy_name", and "versions"
    And the exit code is 0

  # --- US-PM-07: Bind ---

  @pending
  Scenario: Bind an API to a policy
    Given policy "gold" does not include "payments-api"
    When Ravi runs "tyk policy bind --policy gold --api payments-api --versions v1"
    Then the Dashboard receives an update for policy "gold"
    And the updated access_rights includes "m3n4o5p6q7r8" with version "v1"
    And stderr displays "Added payments-api to policy 'gold'"
    And the exit code is 0

  @pending
  Scenario: Bind uses API default version when --versions omitted
    Given policy "gold" does not include "payments-api"
    When Ravi runs "tyk policy bind --policy gold --api payments-api"
    Then the updated access_rights includes "m3n4o5p6q7r8" with the API default version
    And the exit code is 0

  @pending
  Scenario: Bind fails when API is already bound
    Given policy "gold" already includes "users-api"
    When Ravi runs "tyk policy bind --policy gold --api users-api"
    Then stderr displays "API already in policy access list"
    And the exit code is 2

  @pending
  Scenario: Bind fails when policy does not exist
    When Ravi runs "tyk policy bind --policy nonexistent --api users-api"
    Then stderr displays "policy 'nonexistent' not found"
    And the exit code is 3

  @pending
  Scenario: Bind fails when API reference is invalid
    When Ravi runs "tyk policy bind --policy gold --api nonexistent-api"
    Then stderr displays "No API matching 'nonexistent-api' found"
    And the exit code is 3

  # --- US-PM-07: Unbind ---

  @pending
  Scenario: Unbind an API from a policy
    Given policy "gold" includes "orders-api"
    When Ravi runs "tyk policy unbind --policy gold --api orders-api"
    Then the Dashboard receives an update for policy "gold"
    And the updated access_rights does not include "g7h8i9j0k1l2"
    And stderr displays "Removed orders-api from policy 'gold'"
    And the exit code is 0

  @pending
  Scenario: Unbind fails when API is not in the policy
    Given policy "silver" does not include "orders-api"
    When Ravi runs "tyk policy unbind --policy silver --api orders-api"
    Then stderr displays "API not found in policy access list"
    And the exit code is 2

  @pending
  Scenario: Unbind fails when policy does not exist
    When Ravi runs "tyk policy unbind --policy nonexistent --api users-api"
    Then stderr displays "policy 'nonexistent' not found"
    And the exit code is 3
