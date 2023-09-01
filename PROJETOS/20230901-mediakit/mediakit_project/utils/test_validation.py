from .validation import make_dict_validator, validate_object


def test_validator_basic():
    validate_object("core.string", "Teste")
    make_dict_validator(
        "core.test.validator1",
        nome="core.string",
        idade="core.int",
        peso="core.float",
    )
    assert validate_object("core.test.validator1", dict(
        nome="Joao",
        idade="19",
        peso="69.5"
    )) == dict(nome="Joao", idade=19, peso=69.5)


def test_validator_default_dict():
    make_dict_validator(
        "core.test.validator_defaultdict",
        __default__="core.int",
        nome="core.string"
    )
    assert validate_object("core.test.validator_defaultdict", dict(
        a=2,
        b="2",
        c=54.3,
        nome="2"
    )) == dict(a=2, b=2, c=54, nome="2")
