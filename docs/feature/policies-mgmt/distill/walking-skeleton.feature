# Walking Skeleton: Policy Management Vertical Slice
# Proves the full stack: CLI -> policy logic -> client -> Dashboard API (mocked)
# These scenarios are implemented FIRST and enabled for TDD.

Feature: Policy management walking skeleton
  As Ravi Patel, a platform engineer
  I want to list existing policies and apply a new policy from a YAML file
  So that I can verify the complete policy management stack works end-to-end

  Background:
    Given Ravi has a configured environment "staging"
    And the Dashboard has the following APIs:
      | api_id       | name       | listen_path |
      | a1b2c3d4e5f6 | users-api  | /users/     |
      | g7h8i9j0k1l2 | orders-api | /orders/    |

  # --- Walking Skeleton 1: List policies (read path) ---

  @walking_skeleton
  Scenario: Ravi lists policies and sees an empty inventory
    Given the Dashboard has no policies
    When Ravi runs "tyk policy list"
    Then Ravi sees "No policies found." on stderr
    And the exit code is 0

  @walking_skeleton
  Scenario: Ravi lists policies and sees existing policies
    Given the Dashboard has the following policies:
      | _id    | name        | api_count | tags       |
      | gold   | Gold Plan   | 3         | gold, paid |
      | silver | Silver Plan | 2         | silver     |
    When Ravi runs "tyk policy list"
    Then stdout displays a table with columns "ID", "Name", "APIs", "Tags"
    And stdout contains "gold" and "Gold Plan"
    And stdout contains "silver" and "Silver Plan"
    And the exit code is 0

  # --- Walking Skeleton 2: Apply a new policy (write path) ---

  @walking_skeleton
  Scenario: Ravi applies a new policy with name-based selectors
    Given no policy with id "platinum" exists on the Dashboard
    And Ravi has a file "policies/platinum.yaml" with content:
      """
      apiVersion: tyk.tyktech/v1
      kind: Policy
      metadata:
        id: platinum
        name: Platinum Plan
        tags: [platinum, paid]
      spec:
        rateLimit:
          requests: 5000
          per: 1m
        quota:
          limit: 500000
          period: 30d
        keyTTL: 0
        access:
          - name: users-api
            versions: [v1]
      """
    When Ravi runs "tyk policy apply -f policies/platinum.yaml"
    Then the selector "name: users-api" resolves to "a1b2c3d4e5f6"
    And the Dashboard receives a create request with policy id "platinum"
    And the Dashboard payload has "rate" equal to 5000
    And the Dashboard payload has "per" equal to 60
    And the Dashboard payload has "quota_max" equal to 500000
    And the Dashboard payload has "quota_renewal_rate" equal to 2592000
    And stderr displays "Policy applied successfully!"
    And the exit code is 0

  @walking_skeleton
  Scenario: Ravi updates an existing policy idempotently
    Given policy "platinum" already exists on the Dashboard with rate 5000
    And Ravi has a file "policies/platinum.yaml" with rate limit 10000
    When Ravi runs "tyk policy apply -f policies/platinum.yaml"
    Then the Dashboard receives an update request for policy "platinum"
    And the Dashboard payload has "rate" equal to 10000
    And stderr displays "Status: updated"
    And the exit code is 0
