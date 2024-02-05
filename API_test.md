# API test
admin and client API testing notes

## Admin API

### Custom Content

- CreateCustomContent

    | Test Response Code | Test Description                          | Remark |
    |--------------------|-------------------------------------------|--------|
    | code 200           | correct input without substruct           |        |
    |                    | correct input with substruct              |        |
    | code 400           | bad format/wrong type                     |        |
    | code 500           | repeated {app_id, org_id,key combination} |        |

    ```
    {
        "key": "content1",
        "type": "assignment",
        "name": "test assignment",
        "details": "string string string",
        "reference": {
            "name": "string",
            "type": "string",
            "reference_key": "string"
        },
        "linked_content": ["content2"]
    }
    ```

- GetCustomContent

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    | code 400           | bad format/wrong type |        |
    | code 500           | no document exist     |        |

- GetCustomContents

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | no input              |        |
    |                    | single input          |        |
    |                    | comma separated input |        |
    | code 400           | bad format/wrong type |        |

- UpdateCustomContent

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    | code 400           | bad format/wrong type |        |
    | code 500           | no document exist     |        |
    |                    | superior key non-exist|        |


- DeleteCustomContent

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    |                    | not found             |        |

### Custom Unit

- CreateCustomUnit

    | Test Response Code | Test Description                          | Remark |
    |--------------------|-------------------------------------------|--------|
    | code 200           | correct input without substruct           |        |
    |                    | correct input with substruct              |        |
    | code 400           | bad format/wrong type                     |        |
    | code 500           | repeated {app_id, org_id,key combination} |        |

    ```
    {
        "key": "unit1",
        "name": "unit1 test",
        "content": [
            {
                "key": "content1",
                "type": "assignment",
                "name": "test assignment",
                "details": "string string string",
                "reference": {
                    "name": "string",
                    "type": "string",
                    "reference_key": "string"
                },
                "linked_content": ["content2"]
            }
        ],
        "schedule": [
            {
                "name": "schedule item 1",
                "user_content": [
                    {
                        "name": "user content 1",
                        "type": "string",
                        "reference_key": "string",
                        "user_data": {
                            "test_1_score": 97,
                            "homework_3_due": "2023-12-12T22:06:05.021Z"
                        },
                        "date_started": "2023-11-17T22:06:05.021Z"
                    }
                ],
                "duration": 0
            }
        ],
        "schedule_start":0
    }
    ```

- GetCustomUnit

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    | code 400           | bad format/wrong type |        |
    | code 500           | no document exist     |        |

- GetCustomUnits

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | no input              |        |
    |                    | single input          |        |
    |                    | comma separated input |        |
    | code 400           | bad format/wrong type |        |

- UpdateCustomUnit

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    | code 400           | bad format/wrong type |        |
    | code 500           | no document exist     |        |
    |                    | superior key non-exist|        |

```
{
    "key": "unit1",
    "name": "unit1 test",
    "content": [
    {
        "key": "content1",
        "type": "assignment",
        "name": "test assignment",
        "details": "string string string",
        "reference": {
            "name": "string",
            "type": "string",
            "reference_key": "string"
        },
        "linked_content": ["content2"]
    }
  ],
    "schedule": [
        {
            "name": "schedule item 1 name",
            "user_content": [
                {
                    "content_key":"content1",
                    "user_data":
                    {
                        "name": "AAAAA",
                        "score": 95
                    }
                }
            ],
            "duration": 1
        },
        {
            "name": "schedule item 2 name",
            "user_content": [
                {
                    "content_key":"content2",
                    "user_data":
                    {
                        "name": "BBBBB",
                        "score": 95
                    }
                }
            ],
            "duration": 2
        }
    ]
}
```

```
{
    "key": "unit1",
    "name": "unit1 test",
    "content_keys":["content1"],
    "schedule": [
        {
            "name": "schedule item 1 name",
            "user_content": [
                {
                    "content_key":"content1",
                    "user_data":
                    {
                        "name": "AAAAA",
                        "score": 95
                    }
                }
            ],
            "duration": 1
        },
        {
            "name": "schedule item 2 name",
            "user_content": [
                {
                    "content_key":"content2",
                    "user_data":
                    {
                        "name": "BBBBB",
                        "score": 95
                    }
                }
            ],
            "duration": 2
        }
    ]
}
```
- DeleteCustomUnit

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    |                    | not found             |        |

### Custom Module

- CreateCustomModule

    | Test Response Code | Test Description                          | Remark |
    |--------------------|-------------------------------------------|--------|
    | code 200           | correct input without substruct           |        |
    |                    | correct input with substruct              |        |
    | code 400           | bad format/wrong type                     |        |
    | code 500           | repeated {app_id, org_id,key combination} |        |

    ```
    {
        "key": "module1",
        "name": "module1 test",
        "units": [
            {
                "key": "unit1",
                "name": "unit1 test",
                "content": [
                    {
                        "key": "content1",
                        "type": "assignment",
                        "name": "test assignment",
                        "details": "string string string",
                        "reference": {
                            "name": "string",
                            "type": "string",
                            "reference_key": "string"
                        },
                        "linked_content": ["content2"]
                    }
                ],
                "schedule": [
                    {
                        "name": "schedule name",
                        "user_content": [
                            {
                                "name": "content name",
                                "type": "string",
                                "reference_key": "string",
                                "user_data": {
                                    "test_1_score": 97,
                                    "homework_3_due": "2023-12-12T22:06:05.021Z"
                                },
                                "date_started": "2023-11-17T22:06:05.021Z"
                            }
                        ],
                        "duration": 0
                    }
                ]
            }
        ]
    }
    ```

- GetCustomModule

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    | code 400           | bad format/wrong type |        |
    | code 500           | no document exist     |        |

- GetCustomModules

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | no input              |        |
    |                    | single input          |        |
    |                    | comma separated input |        |
    | code 400           | bad format/wrong type |        |

- UpdateCustomModule

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    | code 400           | bad format/wrong type |        |
    | code 500           | no document exist     |        |
    |                    | superior key non-exist|        |

    ```
    {
        "name": "module1 update",
        "unit_keys": ["unit1"]
    }
    ```

- DeleteCustomModule

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    |                    | not found             |        |

### Custom Course

- CreateCustomCourse

    | Test Response Code | Test Description                          | Remark |
    |--------------------|-------------------------------------------|--------|
    | code 200           | correct input without substruct           |        |
    |                    | correct input with substruct              |        |
    | code 400           | bad format/wrong type                     |        |
    | code 500           | repeated {app_id, org_id,key combination} |        |

    ```
    {
        "key": "course1",
        "name": "course1 name",
        "modules": [
            {
                "key": "module1",
                "name": "module1 test",
                "units": [
                    {
                        "key": "unit1",
                        "name": "unit1 test",
                        "content": [
                            {
                                "key": "content1",
                                "type": "assignment",
                                "name": "test assignment",
                                "details": "string string string",
                                "reference": {
                                    "name": "string",
                                    "type": "string",
                                    "reference_key": "string"
                                },
                                "linked_content": ["content2"]
                            }
                        ],
                        "schedule": [
                            {
                                "name": "schedule name",
                                "user_content": [
                                    {
                                        "name": "content name",
                                        "type": "string",
                                        "reference_key": "string",
                                        "user_data": {
                                            "test_1_score": 97,
                                            "homework_3_due": "2023-12-12T22:06:05.021Z"
                                        },
                                        "date_started": "2023-11-17T22:06:05.021Z"
                                    }
                                ],
                                "duration": 0
                            }
                        ]
                    }
                ]
            }
        ]
    }
    ```

- GetCustomCourse

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    | code 400           | bad format/wrong type |        |
    | code 500           | no document exist     |        |

- GetCustomCourses

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | no input              |        |
    |                    | single input          |        |
    |                    | comma separated input |        |
    | code 400           | bad format/wrong type |        |

- UpdateCustomCourse

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    | code 400           | bad format/wrong type |        |
    | code 500           | no document exist     |        |
    |                    | superior key non-exist|        |

    ```
    {
        "name": "course1 update",
        "module_keys": ["module1"]
    }
    ```
- DeleteCustomCourse

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    |                    | not found             |        |
    
## Client API

### User Course

- CreateUserCourse

    | Test Response Code | Test Description                          | Remark |
    |--------------------|-------------------------------------------|--------|
    | code 200           | correct input                             |        |
    | code 400           | no course key in databse                  |        |

- GetUserCourse

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    | code 400           | bad format/wrong type |        |
    | code 500           | no document exist     |        |

- GetUserCourses

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | no input              |        |
    |                    | single input          |        |
    |                    | comma separated input |        |
    | code 400           | bad format/wrong type |        |

- UpdateUserCourseUnitProgress

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    | code 400           | bad format/wrong type |        |
    | code 500           | no document exist     |        |

- DeleteUserCourse

    | Test Response Code | Test Description      | Remark |
    |--------------------|-----------------------|--------|
    | code 200           | correct input         |        |
    |                    | not found             |        |
