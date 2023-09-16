from demo_cython_pytest import native_sum

def test_module():
    assert native_sum(2, 2) == 4
