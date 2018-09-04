FROM registry.datadrivendiscovery.org/uncharted_qntfy/ta3ta2_integration/qntfy_main:beta

WORKDIR /

RUN pip3 install --process-dependency-links --upgrade --upgrade-strategy only-if-needed --exists-action w --editable git+https://gitlab.com/datadrivendiscovery/common-primitives@filter-primitives#egg=common_primitives

WORKDIR /app

