#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHART_DIR="${ROOT_DIR}/charts/tenant-operator"
CRD_SOURCE_DIR="${ROOT_DIR}/config/crd/bases"
CRD_DEST_DIR="${CHART_DIR}/crds"
DIST_DIR="${ROOT_DIR}/dist"

CHART_VERSION="${CHART_VERSION:-}"
APP_VERSION="${APP_VERSION:-$CHART_VERSION}"
CHART_FILE="${CHART_DIR}/Chart.yaml"
PACKAGE_CHART="${PACKAGE_CHART:-true}"
if [[ -n "${PYTHON_BIN:-}" ]]; then
  if ! command -v "${PYTHON_BIN}" >/dev/null 2>&1; then
    echo "error: specified PYTHON_BIN '${PYTHON_BIN}' not found in PATH" >&2
    exit 1
  fi
else
  if command -v python >/dev/null 2>&1; then
    PYTHON_BIN="python"
  elif command -v python3 >/dev/null 2>&1; then
    PYTHON_BIN="python3"
  else
    echo "error: python3 (or python) is required but was not found in PATH" >&2
    exit 1
  fi
fi

if ! command -v helm >/dev/null 2>&1; then
  echo "error: helm is required but was not found in PATH" >&2
  exit 1
fi

if [[ ! -d "${CRD_SOURCE_DIR}" ]]; then
  echo "error: CRD source directory not found: ${CRD_SOURCE_DIR}" >&2
  exit 1
fi

echo ">> Syncing CRDs into chart directory"
mkdir -p "${CRD_DEST_DIR}"
rm -f "${CRD_DEST_DIR}/"*.yaml
cp "${CRD_SOURCE_DIR}/"*.yaml "${CRD_DEST_DIR}/"

if [[ -n "${CHART_VERSION}" ]]; then
  echo ">> Setting chart version to ${CHART_VERSION}"
  if [[ -z "${APP_VERSION}" ]]; then
    APP_VERSION="${CHART_VERSION}"
  fi
  export CHART_FILE CHART_VERSION APP_VERSION
  "${PYTHON_BIN}" - <<'PY'
import os
from pathlib import Path

chart_file = Path(os.environ["CHART_FILE"])
chart_version = os.environ["CHART_VERSION"]
app_version = os.environ["APP_VERSION"]

lines = chart_file.read_text().splitlines()
with chart_file.open("w", encoding="utf-8") as f:
    for line in lines:
        if line.startswith("version:"):
            f.write(f"version: {chart_version}\n")
        elif line.startswith("appVersion:"):
            f.write(f'appVersion: "{app_version}"\n')
        else:
            f.write(line + "\n")
PY
fi

echo ">> Validating chart with helm lint"
helm lint "${CHART_DIR}"

if [[ "${PACKAGE_CHART}" != "false" ]]; then
  echo ">> Packaging chart"
  mkdir -p "${DIST_DIR}"
  rm -f "${DIST_DIR}/tenant-operator-"*.tgz 2>/dev/null || true
  helm package "${CHART_DIR}" --destination "${DIST_DIR}"

  echo ">> Chart package contents:"
  ls -1 "${DIST_DIR}/tenant-operator-"*.tgz
else
  echo ">> Skipping packaging step (PACKAGE_CHART=${PACKAGE_CHART})"
fi
