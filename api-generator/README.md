## How it works

1. API is defined in oapi docs and core files
    - driver/adapter.go & driver/api_* files don't need to be touched
    - auto-generated apis are optional and existing apis work as normal

2. APIHandlerGenerator generates file to route path to handler
    - Gets core function, request body, auth type
    - routeApis in driver/adapter.go uses this file to create the endpoints for each path
    - AUTH_TAGS are the different auth handlers defined in auth.go AUTH types are similar
        - These are same ones used in driver/adapter.go

3. APIRoutesGenerator generates file to call core functions
    - Extracts parameters and request body and passes them to core function

## How to create API

1. Define API in docs
    - Define core function
    - Define request body
        - Add necessary imports (api_routes_generator.py) depending on where request body is defined
        - You can use generated types by oapi for example
    - Define auth type
2. Implement core function
    - CoreFunction(*tokenauth.Claims, param 1, param 2, .. param n)
    - params are passed by value if required and pointers if not
3. Generate docs
    - $ make oapi-gen-docs
4. run main.py
    - $ cd api-generator
    - $ python3 main.py
    - $ make fixfmt

## Adding new paramter type to parse

1. Edit getSchemaFromString & parseArray to define needed parameters (e.g time.Time)
2. Add imports (if needed) to api_routes_generator.py

## Editing doc extentions to define generation
These are "x-" extentions

1. Edit main.py constants
2. Edit constants in driver/adapter.go
3. Change all "x-" extentions in docs