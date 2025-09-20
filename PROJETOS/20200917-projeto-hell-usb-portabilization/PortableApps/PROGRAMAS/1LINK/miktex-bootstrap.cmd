@echo off
shift

set MIKTEX_INSTALL=%BINDIR%latex
mkdir %MIKTEX_INSTALL% 2> nul

miktexsetup ^
	--verbose ^
	--local-package-repository=%MIKTEX_INSTALL%\repo ^
	--package-set=complete ^
	download

miktexsetup ^
	--verbose ^
	--local-package-repository=%MIKTEX_INSTALL%\repo ^
	--portable=%MIKTEX_INSTALL% ^
	--use-registry=no ^
	install
