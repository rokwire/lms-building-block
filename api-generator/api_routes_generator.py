class APIRoutesGenerator:
    def __init__(self, core_functions, imports, tags):
        self.core_functions = core_functions
        self.tags = tags
        self.imports = [
            imports['core'],
            imports['model'],
            imports['utils'] + '\n',
            imports['tokenauth'],
            imports['errors'],
            imports['logutils']
        ]

    def generate(self, destination_path):
        with open(destination_path, 'w') as file:
            data = self.create_header_and_imports_apis()
            data += self.create_handler_struct()
            data += self.create_api_handler_funcs()
            data += self.create_new_handler_instance()
            file.write(data)

    def create_header_and_imports_apis(self):
        data = '// Code generated by api-generator DO NOT EDIT.\n'
        data += 'package web\n\n'
        data += 'import (\n'
        for str in self.imports:
            data += str + '\n'
        return data + ')\n\n'

    def create_handler_struct(self):
        data = '// APIsHandler handles the rest APIs implementation\n'
        data += 'type APIsHandler struct {\n'
        data += 'app *core.Application\n}\n'
        return data

    # def create_param_def_from_schema(self, schema):
    #     if schema['type'] == 'array':
    #         return '[]' + self.create_param_def_from_schema(schema['items'])
    #     if schema['type'] == 'string':
    #         if schema.get('format', '') == 'date-time':
    #             return 'time.Time'
    #         return 'string'
    #     if schema['type'] == 'integer':
    #         return 'int'
    #     if schema['type'] == 'boolean':
    #         return 'bool'

    def create_api_handler_funcs(self):
        data = ''
        for tag in self.tags.keys():
            data += f'\n// {tag}\n\n'
            for core_function in self.core_functions:
                if core_function["tag"] == tag:
                    data_type = core_function["data_type"]
                    request_type = core_function["request_type"]
                    data += f'func (a APIsHandler) {core_function["tag"].lower() + core_function["name"]}'
                    if core_function['auth_type']:
                        data += f'{core_function["handler_prototype"]} {{\n'.replace('{data_type}', data_type).replace('{request_type}', request_type)
                    else:
                        data += f'{core_function["handler_prototype"]} {{\n'.replace('{data_type}', data_type).replace('{request_type}', request_type)

                    interface_prototype = core_function["interface_prototype"]
                    params_start = interface_prototype.find("(")
                    params_end = interface_prototype.find(")")
                    params_string = interface_prototype[params_start + 1:params_end]
                    interface_params = []
                    for param in params_string.split(", "):
                        if not param:
                            continue
                        if param.startswith('claims'):
                            interface_params.append('claims')
                            continue
                        if param.startswith('item'):
                            interface_params.append('*item')
                            continue

                        param_parts = param.split(' ')
                        param_name = param_parts[0]
                        param_api_name = core_function["param_names"][param_name]
                        param_type = param_parts[1]
                        required = not param_type.startswith('*')

                        interface_params.append(param_name)
                        data += f'{param_name}, err := utils.GetValue[{param_type}](params, "{param_api_name}", {str(required).lower()})\n'

                        ret = f'errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("{param_api_name}"), err)'
                        if interface_prototype.count('(') > 1:
                            ret = 'nil, ' + ret
                        data += f'if err != nil {{\nreturn {ret}\n}}\n\n'

                    data += f'return a.app.{core_function["tag"]}.{core_function["name"]}({", ".join(interface_params)})'
                    data += '\n}\n\n'
        return data

    def create_new_handler_instance(self):
        data = '// NewAPIsHandler creates new API handler instance\n'
        data += 'func NewAPIsHandler(app *core.Application) APIsHandler {\n'
        data += 'return APIsHandler{app: app}\n}\n'
        return data