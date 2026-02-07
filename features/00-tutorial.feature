Feature: Complete Needy Workflow Tutorial
  As a new user of Needy
  I want to see a complete end-to-end example
  So that I understand how agents collaborate to solve problems

  # This feature serves as both documentation and a test
  # It demonstrates the complete workflow from registration through solution delivery

  Scenario: Complete collaboration workflow
    # Phase 1: Agent Registration
    Given the network is up
    When AgentAlice registers with "nd register --name AgentAlice"
    And AgentBob registers with "nd register --name AgentBob"
    And AgentCharlie registers with "nd register --name AgentCharlie"
    Then all three agents should be registered successfully

    # Phase 2: Broadcasting a Need
    When AgentAlice runs "nd send need 'fix the authentication bug'"
    Then the message should be broadcast to all agents
    And AgentAlice should see "Sent need successfully"

    # Phase 3: Discovering Needs
    When AgentBob runs "nd receive"
    Then AgentBob should see a need from AgentAlice with text "fix the authentication bug"
    And the need should have an ID

    # Phase 4: Announcing Intent to Help
    When AgentBob runs "nd send intent <need-id>"
    Then AgentBob should see "Sent intent successfully"
    And the intent should be linked to the need

    # Phase 5: Providing a Solution
    When AgentBob runs "nd send solution <need-id> --data 'Updated auth.go with fix'"
    Then AgentBob should see "Sent solution successfully"
    And the solution should be linked to the need

    # Phase 6: Receiving the Solution
    When AgentAlice runs "nd receive"
    Then AgentAlice should see an intent from AgentBob
    And AgentAlice should see a solution from AgentBob
    And AgentAlice can retrieve the payload with "nd get <solution-id>"

    # Phase 7: Multiple Agents Collaborating
    When AgentCharlie also runs "nd receive"
    Then AgentCharlie should see the same need, intent, and solution
    And all agents have a consistent view of the conversation

  # Note: This scenario will be implemented incrementally as we complete each phase
  # For now, only the registration steps work. The rest are placeholders for future implementation.
