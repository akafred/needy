Feature: CLI discovery
  As an AI agent
  I need to understand how the CLI works
  So that I can register and use the system

  Scenario: Learning about registration
    Given the `nd` CLI is available
    And the client in not registered
    When I run the `nd` command
    Then the output should explain that registration is required
    And show the usage for the "register" command
  
  Scenario: Learning about sending
    Given the client is registered
    When I run the `nd register` command
    And registration is successful
    Then the output should explain the mechanics of sending and receiving messages

  Scenario: Learning about getting payloads
    Given there is a need message to receive
    When I run the `nd receive` command
    Then the output should explain the usage of the "get" command and include the id of the message

  Scenario: Learning about the workflow of intents
    Given there is a need message to receive
    When I run the `nd receive` command
    Then the output should explain that the agent must issue a "intent" message referring to the id of the need if it wants to offer a solution

  Scenario: Learning about the workflow of solutions
    Given there is a need 
    When I have issued a "intent" message
    Then the output should explain how to offer a solution by issuing a "solution" message referring to the id of the need

  Scenario: Learning about the limited space for messages
    Given the client is registered
    When I use a to long need message with a payload
    Then the output should explain that the message is too long and that it must contain a message and a payload

  Scenario: Learning about the length of intent messages
    Given the client is registered
    When I use a to intent message that is too long
    Then the output should explain that intents should be short and not have payloads

  Scenario: Learning about the limited space for solution messages
    Given the client is registered
    When I use a to long solution message
    Then the output should explain that the solution message is too long and that it must contain a message and a payload

  Scenario: Learning about the listening with timeout
    Given the client is registered
    When I run the `nd receive` command and there is nothing to receive
    Then the output should explain the mechanics of listening with a timeout