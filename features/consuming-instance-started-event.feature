Feature: Import-Cantabular-Dataset

  Background:
    Given the following recipe with id "recipe-happy-01" is available from dp-recipe-api:
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

  Scenario: Consuming an instance-started event with correct RecipeID, InstanceID and JobID
    When this instance-started event is consumed:
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

    Then these category dimension import events should be produced:

            | DimensionID     | InstanceID        | JobID        |
            | dimension-01    | instance-happy-01 | job-happy-01 |
            | dimension-02    | instance-happy-01 | job-happy-01 |