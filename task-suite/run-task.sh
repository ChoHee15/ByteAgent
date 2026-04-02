#!/usr/bin/env bash
set -euo pipefail

usage() {
	echo "usage: $0 <task-name> <agent-executable> [run-root]" >&2
	echo "example: $0 002-bugfix-greeting /codes/code_agent/dist/code-agent" >&2
}

if [[ $# -lt 2 || $# -gt 3 ]]; then
	usage
	exit 1
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
task_name="$1"
agent_bin="$2"
run_root="${3:-}"
task_dir="${script_dir}/tasks/${task_name}"

if [[ ! -d "${task_dir}" ]]; then
	echo "task not found: ${task_name}" >&2
	exit 1
fi

if [[ ! -x "${agent_bin}" ]]; then
	echo "agent executable is not runnable: ${agent_bin}" >&2
	exit 1
fi

if [[ -n "${run_root}" ]]; then
	mkdir -p "${run_root}"
	run_dir="${run_root}/${task_name}"
	if [[ -e "${run_dir}" ]]; then
		echo "run directory already exists: ${run_dir}" >&2
		exit 1
	fi
	mkdir -p "${run_dir}"
else
	run_dir="$(mktemp -d "/tmp/code-agent-task-${task_name}-XXXXXX")"
fi

workspace_dir="${run_dir}/workspace"
logs_dir="${run_dir}/logs"
prompt_file="${run_dir}/prompt.txt"
result_file="${run_dir}/result.env"

mkdir -p "${logs_dir}"

"${script_dir}/create-workspace.sh" "${task_name}" "${workspace_dir}" >"${logs_dir}/workspace.log" 2>&1
cp "${task_dir}/task.md" "${prompt_file}"

start_time="$(date +%s)"
agent_exit=0

(
	cd "${workspace_dir}"
	CODE_AGENT_UNSAFE_AUTO_APPROVE_BASH_WRITES=1 \
	"${agent_bin}" -- "$(cat "${prompt_file}")"
) >"${logs_dir}/agent.stdout.log" 2>"${logs_dir}/agent.stderr.log" || agent_exit=$?

end_time="$(date +%s)"
duration_sec="$((end_time - start_time))"

verify_exit=0
"${task_dir}/verify.sh" "${workspace_dir}" >"${logs_dir}/verify.stdout.log" 2>"${logs_dir}/verify.stderr.log" || verify_exit=$?

status="pass"
if [[ "${agent_exit}" -ne 0 || "${verify_exit}" -ne 0 ]]; then
	status="fail"
fi

cat >"${result_file}" <<EOF
TASK_NAME=${task_name}
STATUS=${status}
AGENT_BIN=${agent_bin}
RUN_DIR=${run_dir}
WORKSPACE_DIR=${workspace_dir}
PROMPT_FILE=${prompt_file}
AGENT_EXIT_CODE=${agent_exit}
VERIFY_EXIT_CODE=${verify_exit}
DURATION_SEC=${duration_sec}
EOF

echo "task: ${task_name}"
echo "status: ${status}"
echo "run dir: ${run_dir}"
echo "agent exit: ${agent_exit}"
echo "verify exit: ${verify_exit}"
echo "duration sec: ${duration_sec}"

if [[ "${status}" != "pass" ]]; then
	echo "agent stdout: ${logs_dir}/agent.stdout.log"
	echo "agent stderr: ${logs_dir}/agent.stderr.log"
	echo "verify stdout: ${logs_dir}/verify.stdout.log"
	echo "verify stderr: ${logs_dir}/verify.stderr.log"
	exit 1
fi
