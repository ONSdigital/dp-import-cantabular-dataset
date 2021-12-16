Feature: Import-Cantabular-Dataset

  Background:
    Given dp-dataset-api is healthy
    And dp-recipe-api is healthy
    And cantabular server is healthy
    And the following recipe with id "recipe-happy-01" is available from dp-recipe-api:
      """
      {
        "alias": "Cantabular Example 1",
        "format": "cantabular-table",
        "_id": "recipe-happy-01",
        "cantabular_blob": "Example",
        "output_instances": [
          {
            "code_lists": [
              {
                "href": "",
                "id": "dimension-01",
                "name": "Dimension 01"
              },
              {
                "href": "",
                "id": "dimension-02",
                "name": "Dimension 02"
              }
            ]
          }
        ],
        "dataset_id": "cantabular-example-1",
        "editions": [
          "2021"
        ],
        "title": "Example Cantabular Dataset"
      }
      """
    And the following response is available from Cantabular from the codebook "Example" and query "?cats=false&v=dimension-01&dimension-02":
      """
      {
        "dataset": {
          "name": "Example"
        },
        "codebook": [
          {
            "name": "dimension-01",
            "label": "Dimension 01",
            "len": 3,
            "codes": [
              "0",
              "1",
              "2"
            ],
            "labels": [
              "London",
              "Liverpool",
              "Belfast"
            ]
          },
          {
            "name": "dimension-02",
            "label": "Dimension 02",
            "len": 2,
            "codes": [
              "E",
              "N"
            ],
            "labels": [
              "England",
              "Northern Ireland"
            ],
            "mapFrom": [
              {
                "sourceNames": [
                  "dimension-01"
                ],
                "codes": [
                  "E",
                  "",
                  "N"
                ]
              }
            ]
          }
        ]
      }
      """

  Scenario: Consuming an instance-started event with correct RecipeID and InstanceID
    When this instance-started event is queued, to be consumed:
      """
      {
        "RecipeId":       "recipe-happy-01",
        "InstanceId":     "instance-happy-01",
        "JobId":          "job-happy-01",
        "CantabularType": "table"
      }
      """
    And the call to update instance "instance-happy-01" is succesful
    And the call to update job "job-happy-01" is succesful
    And the service starts

    Then these category dimension import events should be produced:
      | DimensionID     | InstanceID        | JobID        | CantabularBlob |
      | dimension-01    | instance-happy-01 | job-happy-01 | Example        |
      | dimension-02    | instance-happy-01 | job-happy-01 | Example        |

  Scenario: Consuming an instance-started event with correct RecipeID and InstanceID
    When this instance-started event is queued, to be consumed:
      """
      {
        "RecipeId":       "recipe-happy-01",
        "InstanceId":     "instance-happy-01",
        "JobId":          "job-happy-02",
        "CantabularType": "table"
      }
      """
    And the call to update instance "instance-happy-01" is succesful
    And the call to update job "job-happy-02" is unsuccesful
    And the service starts

    Then these category dimension import events should be produced:
      | DimensionID     | InstanceID        | JobID        | CantabularBlob |
      | dimension-01    | instance-happy-01 | job-happy-02 | Example        |
      | dimension-02    | instance-happy-01 | job-happy-02 | Example        |

  Scenario: Consuming an instance-started event with correct RecipeID and InstanceID
    When this instance-started event is queued, to be consumed:
      """
      {
        "RecipeId":       "recipe-happy-01",
        "InstanceId":     "instance-happy-01",
        "JobId":          "job-happy-02",
        "CantabularType": "table"
      }
      """
    And the call to update instance "instance-happy-01" is unsuccesful
    And the call to update job "job-happy-02" is unsuccesful
    And the service starts

    Then no category dimension import events should be produced

  Scenario: Consuming an instance-started event with incorrect RecipeID
    When this instance-started event is queued, to be consumed:
      """
      {
        "RecipeId":       "2peofjdkm",
        "InstanceId":     "instance-happy-01",
        "JobId":          "job-happy-01",
        "CantabularType": "table"
      }
      """
    And no recipe with id "2peofjdkm" is available from dp-recipe-api
    And the call to update instance "instance-happy-01" is succesful
    And the call to update job "job-happy-01" is succesful
    And the service starts

    Then no category dimension import events should be produced

  Scenario: Consuming an instance-started event with incorrect InstanceID
    When this instance-started event is queued, to be consumed:
      """
      {
        "RecipeId":       "recipe-happy-01",
        "InstanceId":     "03wiroefld",
        "JobId":          "job-happy-01",
        "CantabularType": "table"
      }
      """
    And the call to update instance "03wiroefld" is unsuccesful
    And the call to update job "job-happy-01" is succesful
    And the service starts

    Then no category dimension import events should be produced
