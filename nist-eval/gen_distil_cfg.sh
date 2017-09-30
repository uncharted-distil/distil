#!/bin/sh

# Generates a kubernetes config file from our template based on a caller supplied eval ID.
# NIST will provide something similar in the eval environment to genereate configs for their
# tasks.
#
# Requires Jinja2 library.

export EVAL_ID="distil-dev"

render_template() {
    python -c "from jinja2 import Template; import sys; print(Template(sys.stdin.read()).render(eval_id=\"$EVAL_ID\"));"
}

cat ./distil-k8-template.yml | render_template > kube.yml
