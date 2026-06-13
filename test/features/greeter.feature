Feature: eat godogs
  Background:
    Given the server started

  Scenario: returns a generic greeting
    Given the server is running
    When I send a GET request to the /helloworld endpoint
    Then I see the message Hello, World!

  Scenario: returns a personalized greeting
    Given the server is running
    When I send a POST request with the body `{"name":"test-name"}` to the /helloworld endpoint
    Then I see the message "Hello, test-name!"

  Scenario: returns a joke from the joke service
    Given the server is running
    When I send a GET request to the /dailyjoke endpoint
    Then I see the message Q: What's a computer's favorite snack?

  Scenario: returns a scuba joke from the joke service
    Given the server is running
    When I send a GET request to the /dailyjoke endpoint
    Then I see the message Q: Hey, wanna hear a joke?