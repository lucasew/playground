from typing import Dict, Any
import logging

logger = logging.getLogger(__name__)

__ALL__ = ["Validator", "VALIDATORS", "entity_type"]


class Validator():
    def validate(self, data):
        """
        Validate a JSON entity

        Raise an error if something is not right
        """
        return data

    def materialize(self, data):
        """
        Solved data about the validator data
        """
        return self.validate(data)


class ValidationException(Exception):
    def __init__(self, message: str, data: Any, exception=None, path="item"):
        super().__init__(message)
        self.path = path
        self.data = data
        self.exception = exception


VALIDATORS: Dict[str, Validator] = {}


def normalize_validator(validator):
    if type(validator) is str:
        return VALIDATORS[validator]
    if issubclass(validator, Validator):
        return validator()
    if isinstance(validator, Validator):
        return validator
    raise ValueError("invalid validator")


def entity_type(name: str):
    logger.debug(_("Registering entity: {entity}").format(entity=name))

    def _handler(handler):
        handler = normalize_validator(handler)
        assert isinstance(handler, Validator)
        VALIDATORS[name] = handler
    return _handler


@entity_type("core.string")
class CoreString(Validator):
    def validate(self, data):
        try:
            return str(data)
        except Exception as e:
            raise ValidationException("can't cast to string",
                                      data,
                                      exception=e)


@entity_type("core.int")
class CoreInt(Validator):
    def validate(self, data):
        try:
            return int(data)
        except Exception as e:
            raise ValidationException("can't cast to int", data, exception=e)


@entity_type("core.float")
class CoreFloat(Validator):
    def validate(self, data):
        try:
            return float(data)
        except Exception as e:
            raise ValidationException("can't cast to float", data, exception=e)


def make_dict_validator(type_name, **kwargs):
    validators = {}
    for (k, v) in kwargs.items():
        validators[k] = normalize_validator(v)
    has_default = kwargs.get('__default__') is not None

    @entity_type(type_name)
    class CustomDictValidator(Validator):
        def validate(self, data):
            ret = {}
            data = dict(data)
            for k in validators.keys():
                if k == "__default__":
                    continue
                validation_path = f'["{k}"]'
                if k not in data:
                    raise ValidationException("missing key",
                                              data,
                                              path=validation_path)

            for (k, v) in data.items():
                validation_path = f'["{k}"]'
                if k == "__default__":
                    continue
                if validators.get(k) is None and not has_default:
                    ret[k] = v
                    continue
                try:
                    validator = validators.get(k)
                    if validator is None and has_default:
                        validator = validators['__default__']
                    if validator is None:
                        continue
                    ret[k] = validator.validate(v)
                except ValidationException as e:
                    e.path = f"{validation_path}{e.path}"
                    raise e
            return ret


def make_enum_validator(type_name, **possible_items):
    possible_items = [str(item) for item in possible_items]

    @entity_type(type_name)
    class CustomEnumValidator(Validator):
        def validate(self, data):
            str_data = validate_object(CoreString, data)
            if not str_data not in possible_items:
                raise ValidationException(
                    "invalid item",
                    data,
                    path=f".enum({', '.join(possible_items)})")
            return str_data


make_dict_validator("core.dict")

make_dict_validator(
    "core.type",
    _type='core.string',
    args='core.dict'
)


def validate_object(validator, data):
    return normalize_validator(validator).validate(data)
