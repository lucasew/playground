{ buildGo122Module
, fetchFromGitHub
}:

buildGo122Module rec {
  pname = "openziti";
  version = "1.1.4";

  subPackages = [ "ziti" ];

  src = fetchFromGitHub {
    owner = "openziti";
    repo = "ziti";
    rev = "v${version}";
    hash = "sha256-AYOXmhj5+fJdkbzQ06gc2zDWxb20hQ1x47y64srWuDA=";
  };
  vendorHash = "sha256-1bLcqGe7g1OfOaCuSe1XgolebKEGEtgAnggLaMJXygA=";
  
  ldflags = [
    "-X github.com/openziti/ziti/common/version.Version=${version}"
  ];
}
