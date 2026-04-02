#!/usr/bin/env bash
set -euo pipefail

usage() {
	echo "usage: $0 <agent-executable> [run-root]" >&2
	echo "example: $0 /codes/code_agent/dist/code-agent" >&2
}

if [[ $# -lt 1 || $# -gt 2 ]]; then
	usage
	exit 1
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
agent_bin="$1"
run_root="${2:-}"

if [[ ! -x "${agent_bin}" ]]; then
	echo "agent executable is not runnable: ${agent_bin}" >&2
	exit 1
fi

if [[ -z "${run_root}" ]]; then
	run_root="$(mktemp -d "/tmp/code-agent-task-suite-runall-XXXXXX")"
else
	mkdir -p "${run_root}"
fi

summary_file="${run_root}/summary.tsv"
printf "task\tstatus\tagent_exit\tverify_exit\tduration_sec\trun_dir\n" >"${summary_file}"

pass_count=0
fail_count=0

mapfile -t tasks < <(find "${script_dir}/tasks" -mindepth 1 -maxdepth 1 -type d -printf "%f\n" | sort)

for task_name in "${tasks[@]}"; do
	echo "=== running ${task_name} ==="
	if "${script_dir}/run-task.sh" "${task_name}" "${agent_bin}" "${run_root}"; then
		:
	else
		:
	fi

	result_file="${run_root}/${task_name}/result.env"
	if [[ ! -f "${result_file}" ]]; then
		echo "missing result file for task: ${task_name}" >&2
		fail_count=$((fail_count + 1))
		printf "%s\tfail\tunknown\tunknown\tunknown\t%s\n" "${task_name}" "${run_root}/${task_name}" >>"${summary_file}"
		continue
	fi

	task_status="$(grep '^STATUS=' "${result_file}" | cut -d= -f2-)"
	agent_exit="$(grep '^AGENT_EXIT_CODE=' "${result_file}" | cut -d= -f2-)"
	verify_exit="$(grep '^VERIFY_EXIT_CODE=' "${result_file}" | cut -d= -f2-)"
	duration_sec="$(grep '^DURATION_SEC=' "${result_file}" | cut -d= -f2-)"
	run_dir="$(grep '^RUN_DIR=' "${result_file}" | cut -d= -f2-)"

	printf "%s\t%s\t%s\t%s\t%s\t%s\n" \
		"${task_name}" "${task_status}" "${agent_exit}" "${verify_exit}" "${duration_sec}" "${run_dir}" \
		>>"${summary_file}"

	if [[ "${task_status}" == "pass" ]]; then
		pass_count=$((pass_count + 1))
	else
		fail_count=$((fail_count + 1))
	fi
done

echo
echo "run root: ${run_root}"
echo "summary: ${summary_file}"
echo "passed: ${pass_count}"
echo "failed: ${fail_count}"

if [[ "${fail_count}" -ne 0 ]]; then
	exit 1
fi

