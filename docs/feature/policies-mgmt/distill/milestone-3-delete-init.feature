# Milestone 3: US-PM-04 (Delete) + US-PM-05 (Init)

Feature: Delete policies and scaffold new policy files
  As Ravi Patel, a platform engineer
  I want to delete obsolete policies and quickly scaffold new ones
  So that I can maintain a clean policy inventory and onboard new policies fast

  Background:
    Given Ravi has a configured environment "staging"
    And the Dashboard has the following policies:
      | _id       | name        | api_count | tags       |
      | gold      | Gold Plan   | 3         | gold, paid |
      | free-tier | Free Plan   | 1         | free       |

  # --- US-PM-04: Delete Policy ---

  @pending
  Scenario: Delete a policy with --yes flag skips confirmation
    When Ravi runs "tyk policy delete free-tier --yes"
    Then the policy "free-tier" is removed from the Dashboard
    And stderr displays "Deleted policy 'free-tier'"
    And the exit code is 0

  @pending
  Scenario: Delete a policy with interactive confirmation accepted
    When Ravi runs "tyk policy delete free-tier"
    Then stderr displays "Are you sure you want to delete policy 'free-tier' (Free Plan)?"
    And stderr displays "This policy controls access for 1 API."
    And Ravi confirms with "y"
    Then stderr displays "Deleted policy 'free-tier'"
    And the exit code is 0

  @pending
  Scenario: Delete cancelled by user
    When Ravi runs "tyk policy delete gold"
    And Ravi enters "n" at the confirmation prompt
    Then stderr displays "Delete operation cancelled"
    And the policy "gold" is not removed from the Dashboard
    And the exit code is 0

  @pending
  Scenario: Delete a non-existent policy
    When Ravi runs "tyk policy delete nonexistent --yes"
    Then stderr displays "policy 'nonexistent' not found"
    And the exit code is 3

  @pending
  Scenario: Delete with JSON output
    When Ravi runs "tyk policy delete free-tier --yes --json"
    Then stdout contains valid JSON with "operation" equal to "deleted"
    And stdout JSON has "policy_id" equal to "free-tier"
    And the exit code is 0

  # --- US-PM-05: Init Policy Scaffold ---

  @pending
  Scenario: Scaffold a new policy file
    When Ravi runs "tyk policy init"
    And Ravi enters "platinum" for Policy ID
    And Ravi enters "Platinum Plan" for Policy Name
    Then a file "policies/platinum.yaml" is created
    And the file contains "apiVersion: tyk.tyktech/v1"
    And the file contains "kind: Policy"
    And the file contains "id: platinum"
    And the file contains "name: Platinum Plan"
    And the file contains placeholder rate limit values
    And the file contains a placeholder access entry
    And stderr displays "Scaffolded: policies/platinum.yaml"
    And the exit code is 0

  @pending
  Scenario: Scaffold warns when file already exists
    Given "policies/gold.yaml" already exists
    When Ravi runs "tyk policy init" with ID "gold"
    Then stderr displays overwrite confirmation prompt
    And Ravi enters "n"
    Then the existing file is not modified
    And the exit code is 0

  @pending
  Scenario: Scaffolded file is syntactically valid for apply
    When Ravi scaffolds "policies/test-plan.yaml" via "tyk policy init"
    Then the generated file is valid YAML
    And the file has correct "apiVersion" and "kind" fields
    And the file structure matches the policy YAML format

  @pending
  Scenario: Init does not require Dashboard connectivity
    Given the Dashboard is unreachable
    When Ravi runs "tyk policy init"
    And Ravi enters "offline-test" for Policy ID
    And Ravi enters "Offline Test" for Policy Name
    Then the scaffold file is created successfully
    And the exit code is 0
