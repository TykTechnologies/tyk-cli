Feature: OAS API Version Management
  As a platform engineer managing API configurations
  I want to manage OAS API versions through the CLI
  So that I can safely evolve APIs without breaking existing consumers

  # --- List Versions ---

  Scenario: List versions for an API with multiple versions
    Given Priya has an API "Petstore" (abc123) with versions "v1", "v2", "v3" and default "v1"
    When Priya runs "tyk api versions list --api-id abc123"
    Then the output shows all 3 versions with "v1" marked as default
    And the exit code is 0

  Scenario: List versions with JSON output
    Given Priya has an API "Petstore" (abc123) with versions "v1", "v2" and default "v1"
    When Priya runs "tyk api versions list --api-id abc123 --json"
    Then stdout contains JSON with "api_id", "default", and "versions" array
    And the exit code is 0

  Scenario: List versions for non-existent API
    Given no API exists with ID "nonexistent"
    When Priya runs "tyk api versions list --api-id nonexistent"
    Then stderr shows "Error: API not found: nonexistent"
    And the exit code is 3

  # --- Get Specific Version ---

  Scenario: Get a specific version's OAS document
    Given Priya has an API "Petstore" (abc123) with version "v2" containing an OAS document
    When Priya runs "tyk api get --api-id abc123 --version-name v2"
    Then stdout contains the OAS document for version "v2"
    And stderr shows metadata including version name, listen path, and upstream URL
    And the exit code is 0

  Scenario: Get a version that does not exist
    Given Priya has an API "Petstore" (abc123) with versions "v1", "v2"
    When Priya runs "tyk api get --api-id abc123 --version-name v99"
    Then stderr shows "Error: version \"v99\" not found for API abc123"
    And stderr shows available versions "v1", "v2"
    And the exit code is 3

  # --- Create Version ---

  Scenario: Create a new version from an OAS file
    Given Priya has an API "Petstore" (abc123) with version "v1"
    And Priya has a valid OAS file "petstore-v2.yaml" with upstream "https://api.petstore.io/v2"
    When Priya runs "tyk api versions create --api-id abc123 -f petstore-v2.yaml --version-name v2"
    Then the version "v2" is created for API abc123
    And stderr shows version details including listen path and upstream
    And stderr shows a hint about switch-default command
    And the exit code is 0

  Scenario: Create a new version and set as default
    Given Priya has an API "Petstore" (abc123) with default "v1"
    And Priya has a valid OAS file "petstore-v2.yaml"
    When Priya runs "tyk api versions create --api-id abc123 -f petstore-v2.yaml --version-name v2 --set-default"
    Then the version "v2" is created and set as default for API abc123
    And the exit code is 0

  Scenario: Create a version with a conflicting name
    Given Priya has an API "Petstore" (abc123) with version "v2" already existing
    When Priya runs "tyk api versions create --api-id abc123 -f petstore-v2.yaml --version-name v2"
    Then stderr shows "Error: version \"v2\" already exists for API abc123"
    And the exit code is 4

  Scenario: Create a version with invalid OAS file
    Given Priya has an API "Petstore" (abc123)
    And Priya has a file "broken.yaml" that is not a valid OAS document
    When Priya runs "tyk api versions create --api-id abc123 -f broken.yaml --version-name v2"
    Then stderr shows a validation error describing the OAS issue
    And the exit code is 2

  Scenario: Create a version with JSON output
    Given Priya has an API "Petstore" (abc123)
    And Priya has a valid OAS file "petstore-v2.yaml"
    When Priya runs "tyk api versions create --api-id abc123 -f petstore-v2.yaml --version-name v2 --json"
    Then stdout contains JSON with "action", "api_id", "version_name", "listen_path", "is_default"
    And the exit code is 0

  # --- Switch Default Version ---

  Scenario: Switch default version
    Given Priya has an API "Petstore" (abc123) with versions "v1" (default), "v2", "v3"
    When Priya runs "tyk api versions switch-default --api-id abc123 --version-name v3"
    Then the default version is switched to "v3"
    And stderr shows previous default "v1" and new default "v3"
    And the exit code is 0

  Scenario: Switch default to non-existent version
    Given Priya has an API "Petstore" (abc123) with versions "v1", "v2"
    When Priya runs "tyk api versions switch-default --api-id abc123 --version-name v99"
    Then stderr shows "Error: version \"v99\" not found for API abc123"
    And stderr lists available versions
    And the exit code is 3

  Scenario: Switch default to already-default version
    Given Priya has an API "Petstore" (abc123) with default "v1"
    When Priya runs "tyk api versions switch-default --api-id abc123 --version-name v1"
    Then stderr shows "Version \"v1\" is already the default"
    And the exit code is 0

  Scenario: Switch default with JSON output
    Given Priya has an API "Petstore" (abc123) with versions "v1" (default), "v2"
    When Priya runs "tyk api versions switch-default --api-id abc123 --version-name v2 --json"
    Then stdout contains JSON with "action", "previous_default", "new_default"
    And the exit code is 0

  # --- Apply Integration ---

  Scenario: Apply a single file as a new version
    Given Priya has an API "Petstore" (abc123)
    And Priya has a valid OAS file "petstore-v3.yaml"
    When Priya runs "tyk api apply -f petstore-v3.yaml --version-name v3 --base-api-id abc123"
    Then version "v3" is created for API abc123
    And the exit code is 0

  Scenario: Batch apply discovers versioned API files
    Given Priya has an API "Petstore" (abc123)
    And a directory "apis/" contains:
      | file                | x-tyk-api-gateway.info.versioning.base_api_id | version_name |
      | petstore-base.yaml  | (none -- this is the base)                     | (none)       |
      | petstore-v2.yaml    | abc123                                         | v2           |
      | petstore-v3.yaml    | abc123                                         | v3           |
    When Priya runs "tyk apply -f apis/"
    Then the base API is updated and versions v2, v3 are created
    And the exit code is 0

  Scenario: Dry-run with version operations
    Given Priya has an API "Petstore" (abc123) with version "v1"
    And Priya has a valid OAS file "petstore-v2.yaml"
    When Priya runs "tyk api apply -f petstore-v2.yaml --version-name v2 --base-api-id abc123 --dry-run"
    Then stderr shows "would create version: v2"
    And no changes are made to the Dashboard
    And the exit code is 0
