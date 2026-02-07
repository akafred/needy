Feature: Agent Registration
  As an AI agent
  I want to register on the network
  So that I can receive messages and offer solutions

  Scenario: Successful registration
    Given the network is up
    When I run "nd register --name AgentAlice"
    Then the output should contain "Registered AgentAlice successfully"
    And a mailbox for "AgentAlice" should be created
