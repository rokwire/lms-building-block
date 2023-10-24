from docs_loader import DocsLoader
from api_func_generator import APIHandlerGenerator
from api_routes_generator import APIRoutesGenerator
from api_interfaces_generator import APIInterfacesGenerator

DOCS_FILEPATH = 'driver/web/docs/gen/def.yaml'

X_CORE_FUNCTION = 'x-core-function'
X_DATA_TYPE = 'x-data-type'
X_AUTH_TYPE = 'x-authentication-type'
X_REQUEST_BODY = 'x-request-body'
X_CONVERSION_FUNCTION = 'x-conversion-function'

API_FUNCTION_DEST_PATH = 'driver/web/adapter_helper.go'
API_ROUTES_DEST_PATH = 'driver/web/apis.go'
API_INTERFACES_DEST_PATH = 'core/interfaces/core_gen.go'

loader = DocsLoader(DOCS_FILEPATH, X_CORE_FUNCTION, X_DATA_TYPE, X_AUTH_TYPE, X_REQUEST_BODY, X_CONVERSION_FUNCTION)

core_functions = loader.get_core_functions()
data_types = loader.get_data_types()
request_bodies = loader.get_request_bodies()

APIHandlerGenerator(core_functions, data_types, request_bodies).generate(API_FUNCTION_DEST_PATH)
APIRoutesGenerator(core_functions).generate(API_ROUTES_DEST_PATH)
APIInterfacesGenerator(core_functions).generate(API_INTERFACES_DEST_PATH)