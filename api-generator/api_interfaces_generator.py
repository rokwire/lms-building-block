IMPORTS = [
	'"lms/core/model"\n',
	'"github.com/rokwire/core-auth-library-go/v3/tokenauth"',
]
INTERFACE_TAGS = {'Default': 'default', 'Client': 'client', 'Admin': 'administrative'} # , 'BBs': 'building block', 'TPS': 'third-party service', 'System': 'system'}

class APIInterfacesGenerator:
    def __init__(self, core_functions):
        self.core_functions = core_functions

    def generate(self, destination_path):
        data = ''
        with open(destination_path, 'w') as file:
            data += self.create_header_and_imports()
            data += self.create_interfaces()
            file.write(data)

    def create_header_and_imports(self):
        data = '// Code generated by api-generator DO NOT EDIT.\n'
        data += 'package interfaces\n\n'
        data += 'import (\n'
        for str in IMPORTS:
            data += str + '\n'
        return data + ')\n'

    def create_interfaces(self):
        data = ''
        for tag, comment in INTERFACE_TAGS.items():
            data += f'// {tag} exposes {comment} APIs to the driver adapters\n'
            data += f'type {tag} interface {{'

            data_types = []
            for core_function in self.core_functions:
                if core_function['tag'] == tag:
                    data_type = core_function["data_type"]
                    if data_types.count(data_type) == 0:
                        data += f'\n// {data_type}\n\n'
                        data_types.append(data_type)
                    if core_function['auth_type']:
                        data += f'{core_function["name"]}{core_function["interface_prototype"]}\n'.replace('{data_type}', data_type)
                    else:
                        data += f'{core_function["name"]}{core_function["interface_prototype"]}\n'.replace('{data_type}', data_type)
            data += '}\n\n'
        return data
