class Utils:
    @staticmethod
    def get_camel_case_from_snake(val: str, all_caps: list) -> str:
        snake_split = val.replace('-', "_").split('_')
        return snake_split[0] + ''.join(part.upper() if part in all_caps else part.title() for part in snake_split[1:])