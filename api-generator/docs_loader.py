import yaml
from utils import Utils

class DocsLoader:
    camel_caps = ['id']

    def __init__(self, file_path, gen_types_package, docs_ext_core_function_key, docs_ext_data_type_key, docs_ext_auth_type_key, docs_ext_conv_function_key):
        self.docs = self.load_docs(file_path)
        self.gen_types_package = gen_types_package
        self.docs_ext_core_function_key = docs_ext_core_function_key
        self.docs_ext_data_type_key = docs_ext_data_type_key
        self.docs_ext_auth_type_key = docs_ext_auth_type_key
        self.docs_ext_conv_function_key = docs_ext_conv_function_key
    
    def load_docs(self, file_path):
        with open(file_path) as file:
            return yaml.safe_load(file)
    
    def get_tags(self):
        return [tag['name'] for tag in self.docs['tags']]
    
    def get_data_types(self):
        data_types = set()
        for methods in self.docs['paths'].values():
            for method_data in methods.values():
                if self.docs_ext_data_type_key in method_data:
                    data_types.add(method_data[self.docs_ext_data_type_key])
        return data_types

    def get_request_bodies(self):
        request_bodies = set()
        for methods in self.docs['paths'].values():
            for method_data in methods.values():
                if method_data.get('requestBody', None):
                    try:
                        request_body_ref, request_body = self.get_request_body_type(method_data)
                    except:
                        print(f'{method_data[self.docs_ext_core_function_key]}: failed to parse request body from docs')
                        continue
                    
                    try:
                        request_body_name = request_body.split(',')[-1]
                        data_type_name = method_data[self.docs_ext_data_type_key].split(',')[-1]
                        if request_body_name != data_type_name:
                            request_bodies.add((request_body_ref, request_body, method_data[self.docs_ext_data_type_key]))
                    except:
                        pass
        return request_bodies

    def get_core_functions(self):
        core_functions = []
        for methods in self.docs['paths'].values():
            for method, method_data in methods.items():
                if self.docs_ext_core_function_key in method_data:
                    for tag in method_data['tags']:
                        request_type = data_type = method_data.get(self.docs_ext_data_type_key, None)
                        auth_type = method_data.get(self.docs_ext_auth_type_key, None)
                        handler_prototype = self.get_api_handler_prototype(method, method_data)
                        interface_prototype = handler_prototype.replace('item *{request_type}', 'item {request_type}')

                        parameters = method_data.get('parameters', [])
                        param_prototype, param_names = self.get_core_interface_param_prototype(parameters)
                        if param_prototype:
                            interface_prototype = interface_prototype.replace('params map[string]interface{}', param_prototype)
                        else:
                            interface_prototype = interface_prototype.replace('params map[string]interface{}', '').replace(', ,', ',').replace(', )', ')')
                            
                        if method_data.get('requestBody', None):
                            try:
                                _, request_body = self.get_request_body_type(method_data)
                                if request_body.startswith('model.'):
                                    request_type = request_body
                            except:
                                pass
                        elif method == 'post' or method == 'put':
                            item_param_start = interface_prototype.index(', item')
                            item_param_end = interface_prototype.index(')', item_param_start)
                            interface_prototype = interface_prototype[:item_param_start] + interface_prototype[item_param_end:]
                        if not auth_type:
                            interface_prototype = interface_prototype.replace('claims *tokenauth.Claims', '').replace('( ,', '(')

                        core_functions.append({
                            'name': method_data[self.docs_ext_core_function_key],
                            'tag': tag,
                            'handler_prototype': handler_prototype,
                            'interface_prototype': interface_prototype,
                            'param_names': param_names,
                            'data_type': data_type,
                            'request_type': request_type,
                            'auth_type': auth_type,
                            'conv_function': method_data.get(self.docs_ext_conv_function_key, None)
                        })
        return core_functions
    
    def get_api_handler_prototype(self, method, method_data):
        if method == 'get':
            try:
                many = method_data['responses']['200']['content']['application/json']['schema']['type'] == 'array'
            except:
                many = False
            if many:
                return '(claims *tokenauth.Claims, params map[string]interface{}) ([]{data_type}, error)'
            else:
                return '(claims *tokenauth.Claims, params map[string]interface{}) (*{data_type}, error)'
        elif method == 'post' or method == 'put':
            return '(claims *tokenauth.Claims, params map[string]interface{}, item *{request_type}) (*{data_type}, error)'
        elif method == 'delete':
            return '(claims *tokenauth.Claims, params map[string]interface{}) error'
        return ''
    
    def get_core_interface_param_prototype(self, parameters):
        param_prototype = ''
        param_names = {}
        for i, param in enumerate(parameters):
            param_type = self.get_param_arg_from_param(param['schema'])
            if param_type:
                if i > 0:
                    param_prototype += ', '

                param_name = param["name"]
                camel_name = Utils.get_camel_case_from_snake(param_name, self.camel_caps)
                param_prototype += f'{camel_name} '
                param_names[camel_name] = param["name"]
                
                if not param.get("required", False):
                    param_prototype += '*'
                param_prototype += param_type
        return param_prototype, param_names
    
    def get_param_arg_from_param(self, param):
        if param['type'] == 'array':
            return '[]' + self.get_param_arg_from_param(param['items'])
        elif param['type'] == 'string':
            if param.get('format', '') == 'date-time':
                return 'time.Time'
            return 'string'
        elif param['type'] == 'integer':
            return 'int'
        elif param['type'] == 'number':
            return 'float64'
        elif param['type'] == 'boolean':
            return 'bool'
        return ''
    
    def get_request_body_type(self, method_data):
        request_body_ref = method_data['requestBody']['content']['application/json']['schema']['$ref']

        request_body_ref_parts = request_body_ref.split('/')
        request_body_snake = request_body_ref_parts[-1]
        request_body = Utils.get_camel_case_from_snake(request_body_snake, self.camel_caps)
        if method_data.get(self.docs_ext_conv_function_key, None):
            if self.gen_types_package != 'web':
                return request_body_ref, f'{self.gen_types_package}.{request_body[0].upper() + request_body[1:]}'
            else:
                return request_body_ref, request_body[0].lower() + request_body[1:]
        return request_body_ref, 'model.' + request_body