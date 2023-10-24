import yaml

class DocsLoader:
    def __init__(self, file_path, docs_ext_core_function_key, docs_ext_data_type_key, docs_ext_auth_type_key, docs_ext_request_body_key, docs_ext_conv_function_key):
        self.docs = self.load_docs(file_path)
        self.docs_ext_core_function_key = docs_ext_core_function_key
        self.docs_ext_data_type_key = docs_ext_data_type_key
        self.docs_ext_auth_type_key = docs_ext_auth_type_key
        self.docs_ext_request_body_key = docs_ext_request_body_key
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
                if self.docs_ext_request_body_key in method_data:
                    request_bodies.add((method_data[self.docs_ext_request_body_key], method_data[self.docs_ext_data_type_key]))
        return request_bodies

    def get_core_functions(self):
        core_functions = []
        for methods in self.docs['paths'].values():
            for method, method_data in methods.items():
                if self.docs_ext_core_function_key in method_data:
                    for tag in method_data['tags']:
                        auth_type = method_data.get(self.docs_ext_auth_type_key, None)
                        handler_prototype = self.get_api_handler_prototype(method, method_data)

                        parameters = method_data.get('parameters', [])
                        param_prototype, param_names = self.get_core_interface_prototype(parameters)
                        if param_prototype:
                            interface_prototype = handler_prototype.replace('params map[string]interface{}', param_prototype)
                        else:
                            interface_prototype = handler_prototype.replace('params map[string]interface{}', '').replace(', ,', ',').replace(', )', ')')
                        if not auth_type:
                            interface_prototype = interface_prototype.replace('claims *tokenauth.Claims', '').replace('( ,', '(')

                        core_functions.append({
                            'name': method_data[self.docs_ext_core_function_key],
                            'tag': tag,
                            'handler_prototype': handler_prototype,
                            'interface_prototype': interface_prototype,
                            'param_names': param_names,
                            'data_type': method_data.get(self.docs_ext_data_type_key, None),
                            'auth_type': auth_type,
                            'request_body': method_data.get(self.docs_ext_request_body_key, None),
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
            return '(claims *tokenauth.Claims, params map[string]interface{}, item {data_type}) (*{data_type}, error)'
        elif method == 'delete':
            return '(claims *tokenauth.Claims, params map[string]interface{}) error'
        return ''
    
    def get_core_interface_prototype(self, parameters):
        param_prototype = ''
        param_names = {}
        for i, param in enumerate(parameters):
            param_type = self.get_param_arg_from_param(param['schema'])
            if param_type:
                if i > 0:
                    param_prototype += ', '

                param_name = param["name"].replace('-', "_")
                snake_split = str(param_name).split('_')
                camel_name = snake_split[0] + ''.join(part.upper() if part == "id" else part.title() for part in snake_split[1:])
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