Feature: Import-Cantabular-Dataset-Unhealthy

  Background:
    Given dp-dataset-api is unhealthy
    And dp-recipe-api is healthy
    And cantabular server is healthy
    And cantabular api extension is healthy

  Scenario: Not consuming instance-started events, because a dependency is not healthy
    When the service starts
    And this instance-started event is queued, to be consumed:
      """
      {
        "RecipeId":       "recipe-happy-01",
        "InstanceId":     "instance-happy-01",
        "JobId":          "job-happy-01",
        "CantabularType": "table"
      }
      """

    Then no category dimension import events should be produced
