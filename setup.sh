#!/bin/bash

set -euo pipefail

pre-commit install -f  -t commit-msg
pre-commit install -f -t pre-commit
