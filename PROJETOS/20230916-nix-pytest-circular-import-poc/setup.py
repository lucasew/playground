from distutils.core import Extension
from setuptools import setup, find_packages
from Cython.Build import cythonize
from Cython.Distutils import build_ext
# from distutils.command.build_ext import build_ext


def ext_modules():
    modules = []
    includes = []
    libraries = []
    modules += cythonize(Extension(
        "*",
        ["demo_cython_pytest/module.pyx"],
        include_dirs=includes,
        libraries=libraries
    ))
    return modules

setup(
    name="demo_cython_pytest",
    version="0.0.1",
    description="Circular imports?",
    author="lucasew",
    packages=["demo_cython_pytest"],
    ext_modules=ext_modules(),
    include_package_data=True,
    cmdclass=dict(
        build_ext=build_ext
    )
)
