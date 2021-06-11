Feature: Helloworld



  Scenario: Posting and checking a response
    When these hello events are consumed:
            | CantabularType | 
            | table          |
    Then I should receive a hello-world response