{ buildPythonPackage
, pytestCheckHook
}:

buildPythonPackage {
  pname = "mediakit_project";
  version = builtins.readFile ./mediakit_project/VERSION;
  src = ./.;

  propagatedBuildInputs = [
  ];

  checkInputs = [ pytestCheckHook ];

  pythonImportsCheck = [ "mediakit_project" ];
}
