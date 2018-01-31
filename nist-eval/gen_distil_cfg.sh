#!/bin/bash

# Generates a kubernetes config file from our template based on a caller supplied eval ID.
# NIST will provide something similar in the eval environment to genereate configs for their
# tasks.
#
# Requires Jinja2 library, jinja2-cli:
#
# pip install jinja2 jinja2-cli
jinja2 ./distil-k8-template.yml local_data.json > kube.yml
