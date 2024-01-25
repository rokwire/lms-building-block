from docs_loader import DocsLoader
from api_func_generator import APIHandlerGenerator
from api_routes_generator import APIRoutesGenerator
from api_interfaces_generator import APIInterfacesGenerator

MODULE_NAME = 'lms'
DOCS_FILEPATH = 'driver/web/docs/gen/def.yaml'
GEN_TYPES_PACKAGE = 'Def'

IMPORTS = {
    'core': f'"{MODULE_NAME}/core"',
    'model': f'"{MODULE_NAME}/core/model"',
    GEN_TYPES_PACKAGE: f'{GEN_TYPES_PACKAGE} "{MODULE_NAME}/driver/web/docs/gen"',
    'utils': f'"{MODULE_NAME}/utils"',
    'mux': '"github.com/gorilla/mux"',
    'tokenauth': '"github.com/rokwire/core-auth-library-go/v3/tokenauth"',
    'errors': '"github.com/rokwire/logging-library-go/v2/errors"',
    'logutils': '"github.com/rokwire/logging-library-go/v2/logutils"'
}
TAGS = {'Default': 'default', 'Client': 'client', 'Admin': 'administrative'} # , 'BBs': 'building block', 'TPS': 'third-party service', 'System': 'system'}
AUTH_HANDLERS_TYPES = ['User', 'Standard', 'Authenticated', 'Permissions']

X_CORE_FUNCTION = 'x-core-function'
X_DATA_TYPE = 'x-data-type'
X_AUTH_TYPE = 'x-authentication-type'
X_CONVERSION_FUNCTION = 'x-conversion-function'

API_FUNCTION_DEST_PATH = 'driver/web/adapter_helper.go'
API_ROUTES_DEST_PATH = 'driver/web/apis.go'
API_INTERFACES_DEST_PATH = 'core/interfaces/core.go'

loader = DocsLoader(DOCS_FILEPATH, GEN_TYPES_PACKAGE, X_CORE_FUNCTION, X_DATA_TYPE, X_AUTH_TYPE, X_CONVERSION_FUNCTION)

core_functions = loader.get_core_functions()
data_types = loader.get_data_types()
request_bodies = loader.get_request_bodies()

APIHandlerGenerator(core_functions, data_types, request_bodies, IMPORTS, list(TAGS.keys())[1:], AUTH_HANDLERS_TYPES, GEN_TYPES_PACKAGE).generate(API_FUNCTION_DEST_PATH)
APIRoutesGenerator(core_functions, IMPORTS, TAGS).generate(API_ROUTES_DEST_PATH)
APIInterfacesGenerator(core_functions, IMPORTS, TAGS).generate(API_INTERFACES_DEST_PATH)