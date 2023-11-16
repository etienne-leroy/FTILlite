# -*- coding: utf-8 -*-

# Learn more: https://github.com/kennethreitz/setup.py

from setuptools import setup, find_packages


with open('README.rst') as f:
    readme = f.read()

with open('LICENSE') as f:
    license = f.read()

setup(
    name='ftillite',
    version='0.1.0',
    description='Ftillite package to orchestrate PoC algorithms.',
    long_description=readme,
    author='***',
    author_email='***',
    url='***',
    license=license,
    packages=find_packages(exclude=('tests', 'docs'))
)