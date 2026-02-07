Feature: Agent Registration
  As an AI agent
  I want to register on the network
  So that I can receive messages and offer solutions

  Scenario: Successful registration
    Given the network is up
    When I run "nd register --name AgentAlice"
    Then the output should contain "Registered AgentAlice successfully"
    And a mailbox for "AgentAlice" should be created

  Scenario: Registration fails when network is down
    Given the network is not running
    When I run "nd register --name AgentBob"
    Then the output should contain "Error: Could not connect to network"
    And the command should fail

  Scenario: Impersonation is prevented
    Given the network is up
    And "AgentAlice" is already registered from a different client
    When I run "nd register --name AgentAlice"
    Then the output should contain "Error: Agent name 'AgentAlice' is already registered"
    And the command should fail

  Scenario: Re-registration by same client succeeds
    Given the network is up
    And I previously registered as "AgentAlice"
    When I run "nd register --name AgentAlice"
    Then the output should contain "Re-registered AgentAlice successfully"
    And the mailbox for "AgentAlice" should be reconnected

  Scenario: Registration without name flag
    Given the network is up
    When I run "nd register"
    Then the output should contain "Error: --name flag is required"
    And the command should fail

  Scenario: Multiple agents can register
    Given the network is up
    When I run "nd register --name AgentAlice"
    And I run "nd register --name AgentBob"
    And I run "nd register --name AgentCharlie"
    Then all registrations should succeed
    And mailboxes for "AgentAlice", "AgentBob", and "AgentCharlie" should exist
