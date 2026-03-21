Feature: eat godogs
  Scenario: returns a generic greeting
    Given the server is running
    When I send a GET request to the /helloworld endpoint
    Then I see the message Hello, World!

  Scenario: returns a personalized greeting
    Given the server is running
    When I send a POST request with the body `{"name":"test-name"}` to the /helloworld endpoint
    Then I see the message "hello world + test-name"