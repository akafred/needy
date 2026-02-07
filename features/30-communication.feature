Feature: Agent Communication
  As an AI agent
  I want to send needs and receive solutions
  So that I can collaborate with other agents

  Scenario: Sending a need and receiving it
    Given a registered agent "AgentAlice"
    And a registered agent "AgentBob"
    When agent "AgentAlice" runs "nd send need 'fix the bug' --data 'bug details...'"
    Then agent "AgentBob" should receive a message with text "fix the bug"

  Scenario: Intent must precede solution
    Given a registered agent "AgentAlice"
    And a registered agent "AgentBob"
    And agent "AgentAlice" has sent a need "fix the bug"
    When agent "AgentBob" runs "nd send solution 'fixed it'"
    Then the command should fail with "You must first announce intent to respond"

  Scenario: Successful solution flow
    Given a registered agent "AgentAlice"
    And a registered agent "AgentBob"
    And agent "AgentAlice" has sent a need "fix the bug"
    When agent "AgentBob" runs "nd send intent 'fix the bug'"
    And agent "AgentBob" runs "nd send solution 'fix the bug' 'fixed it'"
    Then the command should succeed
