#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
	echo "usage: $0 <task-name> <destination>" >&2
	exit 1
fi

task_name="$1"
destination="$2"
task_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/tasks/${task_name}"
template_dir="${task_dir}/workspace-template"

if [[ ! -d "${task_dir}" ]]; then
	echo "task not found: ${task_name}" >&2
	exit 1
fi

if [[ ! -d "${template_dir}" ]]; then
	echo "workspace template not found for task: ${task_name}" >&2
	exit 1
fi

if [[ -e "${destination}" ]]; then
	echo "destination already exists: ${destination}" >&2
	exit 1
fi

mkdir -p "$(dirname "${destination}")"
cp -R "${template_dir}" "${destination}"
echo "workspace created at ${destination}"
