class APIHandlerGenerator:
    def __init__(self, core_functions, data_types, request_bodies, imports, auth_handlers, auth_handlers_types, gen_types_package):
        self.core_functions = core_functions
        self.data_types = data_types
        self.request_bodies = request_bodies
        self.imports = [
            imports['model'],
            imports[gen_types_package] + '\n',
            imports['mux'] + '\n',
            imports['tokenauth'],
            imports['errors'],
            imports['logutils']
        ]
        self.auth_handlers = auth_handlers
        self.auth_handlers_types = auth_handlers_types

    def generate(self, destination_path):
        with open(destination_path, 'w') as file:
            data = self.create_header_and_imports()
            data += self.create_api_data_type_interface()
            data += self.create_request_data_type_interface()
            data += self.create_register_api_handler()
            data += self.create_get_auth_handler()
            data += self.create_get_core_handler()
            data += self.create_get_conv_func()

            file.write(data)

    def create_header_and_imports(self):
        data = '// Code generated by api-generator DO NOT EDIT.\n'
        data += 'package web\n\n'
        data += 'import (\n'
        for str in self.imports:
            data += str + '\n'
        return data + ')\n'
    
    def create_api_data_type_interface(self):
        data = '// apiDataType represents any stored data type that may be read/written by an API\n'
        data += 'type apiDataType interface {\n'
        for i, data_type in enumerate(self.data_types):
            if i > 0:
                data += f' |\n{data_type}'
            else:
                data += data_type
        return data + '\n}\n\n'
    
    def create_request_data_type_interface(self):
        data = '// requestDataType represents any data type that may be sent in an API request body\n'
        data += 'type requestDataType interface {\n'
        data += 'apiDataType'
        if len(self.request_bodies) > 0:
            data += '|\n'
        for i, request_body in enumerate(self.request_bodies):
            if data.count(request_body[1]) == 0:
                if i > 0:
                    data += f' |\n{request_body[1]}'
                else:
                    data += request_body[1]

        return data + '\n}\n\n'

    def create_register_api_handler(self):
        data = 'func (a *Adapter) registerHandler(router *mux.Router, pathStr string, method string, tag string, coreFunc string, dataType string, authType interface{},\n \
            requestBody interface{}, conversionFunc interface{}) error {'
        data += 'authorization, err := a.getAuthHandler(tag, authType)\n'
        data += 'if err != nil {\nreturn errors.WrapErrorAction(logutils.ActionGet, "api auth handler", nil, err)\n}\n\n'

        data += 'coreHandler, err := a.getCoreHandler(tag, coreFunc)\n'
        data += 'if err != nil {\nreturn errors.WrapErrorAction(logutils.ActionGet, "api core handler", nil, err)\n}\n\n'

        if len(self.request_bodies) > 0:
            data += 'var convFunc interface{}\nif conversionFunc != nil{\nconvFunc, err = a.getConversionFunc(conversionFunc)\n'
            data += 'if err != nil {\nreturn errors.WrapErrorAction(logutils.ActionGet, "request body conversion function", nil, err)\n}\n}\n\n'
        
        data += 'switch dataType {\n'
        for data_type in self.data_types:
            message_data_type = 'logutils.MessageDataType(dataType)'
            data_type_parts = data_type.split('.')
            if len(data_type_parts) == 2:
                message_data_type = '.'.join([data_type_parts[0], 'Type' + data_type_parts[1]])
            data += f'case "{data_type}":\n'

            request_cases = ''
            for request_body in self.request_bodies:
                if request_body[2] == data_type:
                    request_cases += f'case "{request_body[0]}":\n'
                    use_conv = not request_body[1].startswith('model.')
                    generic_params = f'{data_type}, {data_type}, {request_body[1]}'
                    conv_arg = ''
                    if use_conv:
                        request_cases += f'convert, ok := convFunc.(func(*tokenauth.Claims, *{request_body[1]}) (*{data_type}, error))\n'
                        request_cases += 'if !ok {\nreturn errors.ErrorData(logutils.StatusInvalid, "request body conversion function", &logutils.FieldArgs{"x-conversion-function": conversionFunc})\n}\n\n'
                        generic_params = f'{data_type}, {request_body[1]}, {data_type}'
                        conv_arg = 'conversionFunc: convert, '
                    request_cases += f'handler := apiHandler[{generic_params}]{{authorization: authorization, {conv_arg}messageDataType: {message_data_type}}}\n'
                    request_cases += f'err = setCoreHandler[{generic_params}](&handler, coreHandler, method, tag, coreFunc)\n'
                    request_cases += 'if err != nil {\nreturn errors.WrapErrorAction(logutils.ActionApply, "api core handler", &logutils.FieldArgs{"name": tag+"."+coreFunc}, err)\n}\n\n'
                    request_cases += f'router.HandleFunc(pathStr, handleRequest[{generic_params}](&handler, a.paths, a.logger)).Methods(method)\n'
            if request_cases != '':
                request_cases += f'default:\nhandler := apiHandler[{data_type}, {data_type}, {data_type}]{{authorization: authorization, messageDataType: {message_data_type}}}\n'
                request_cases += f'err = setCoreHandler[{data_type}, {data_type}, {data_type}](&handler, coreHandler, method, tag, coreFunc)\n'
                request_cases += 'if err != nil {\nreturn errors.WrapErrorAction(logutils.ActionApply, "api core handler", &logutils.FieldArgs{"name": tag+"."+coreFunc}, err)\n}\n\n'
                request_cases += f'router.HandleFunc(pathStr, handleRequest[{data_type}, {data_type}, {data_type}](&handler, a.paths, a.logger)).Methods(method)\n'
                data += f'switch requestBody {{\n{request_cases}}}\n'
            else:
                data += f'handler := apiHandler[{data_type}, {data_type}, {data_type}]{{authorization: authorization, messageDataType: {message_data_type}}}\n'
                data += f'err = setCoreHandler[{data_type}, {data_type}, {data_type}](&handler, coreHandler, method, tag, coreFunc)\n'
                data += 'if err != nil {\nreturn errors.WrapErrorAction(logutils.ActionApply, "api core handler", &logutils.FieldArgs{"name": tag+"."+coreFunc}, err)\n}\n\n'
                data += f'router.HandleFunc(pathStr, handleRequest[{data_type}, {data_type}, {data_type}](&handler, a.paths, a.logger)).Methods(method)\n'
        data += f'default:\nreturn errors.ErrorData(logutils.StatusInvalid, "data type reference", nil)\n}}\n\n'
        
        return data + 'return nil\n}\n\n'
    
    def create_get_auth_handler(self):
        data = 'func (a *Adapter) getAuthHandler(tag string, ref interface{}) (tokenauth.Handler, error) {\n'
        data += 'if ref == nil {\n'
        data += 'return nil, nil\n}\n\n'
        data += 'var handler tokenauth.Handlers\n'
        data += 'switch tag {\n'
        for tag in self.auth_handlers:
            data += f'case "{tag}":\n'
            data += f'handler = a.auth.{tag.lower()}\n'
        data += f'default:\nreturn nil, errors.ErrorData(logutils.StatusInvalid, "tag", &logutils.FieldArgs{{"tag": tag}})}}\n\n'
        data += 'switch ref {\n'
        for auth_type in self.auth_handlers_types:
            data += f'case "{auth_type}":\n'
            data += f'return handler.{auth_type}, nil\n'
        data += f'default:\nreturn nil, errors.ErrorData(logutils.StatusInvalid, "authentication type reference", &logutils.FieldArgs{{"ref": ref}})}}\n}}\n\n'
        return data
    
    def create_get_core_handler(self):
        data = 'func (a *Adapter) getCoreHandler(tag string, ref string) (interface{}, error) {\n'
        data += 'switch tag + ref {\n'
        for core_function in self.core_functions:
            tag = core_function.get('tag', None)
            name = core_function.get('name', None)
            if tag and name:
                data += f'case "{tag + name}":\n'
                data += f'return a.apisHandler.{tag.lower() + name}, nil\n'
        return data + 'default:\nreturn nil, errors.ErrorData(logutils.StatusInvalid, "core function", logutils.StringArgs(tag + ref))}\n}\n\n'
    
    def create_get_conv_func(self):
        data = 'func (a *Adapter) getConversionFunc(ref interface{}) (interface{}, error) {\n'
        data += 'if ref == nil {\nreturn nil, nil\n}\n\n'
        data += 'switch ref {\n'
        for core_function in self.core_functions:
            conv_function = core_function.get('conv_function', None)
            if conv_function and data.count(conv_function) == 0:
                data += f'case "{conv_function}":\n'
                data += f'return {conv_function}, nil\n'
        return data + f'default:\nreturn nil, errors.ErrorData(logutils.StatusInvalid, "conversion function reference", &logutils.FieldArgs{{"ref": ref}})}}\n}}\n'