{ buildPythonPackage
, pytestCheckHook
, cython
, python
}:

buildPythonPackage {
  pname = "demo_cython_pytest";
  version = "0.0.1";

  src = ./.;

  pythonImportsCheck = [ "demo_cython_pytest.module.native_sum" ];

  nativeBuildInputs = [
    # pytestCheckHook
    cython
  ];
}
