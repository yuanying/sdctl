#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
skills_dir="${repo_root}/skills"

targets=(
  "${HOME}/.agents/skills"
  "${HOME}/.claude/skills"
)

if [[ ! -d "${skills_dir}" ]]; then
  echo "skills directory not found: ${skills_dir}" >&2
  exit 1
fi

skill_paths=()
for skill_path in "${skills_dir}"/*; do
  [[ -d "${skill_path}" ]] || continue
  [[ -f "${skill_path}/SKILL.md" ]] || continue
  skill_paths+=("${skill_path}")
done

if [[ ${#skill_paths[@]} -eq 0 ]]; then
  echo "no skills found in ${skills_dir}" >&2
  exit 1
fi

has_conflict=0

for target_dir in "${targets[@]}"; do
  for skill_path in "${skill_paths[@]}"; do
    skill_name="$(basename "${skill_path}")"
    link_path="${target_dir}/${skill_name}"

    if [[ -L "${link_path}" ]]; then
      current_target="$(readlink "${link_path}")"
      if [[ "${current_target}" == "${skill_path}" ]]; then
        continue
      fi
      echo "conflict: ${link_path} is a symlink to ${current_target}" >&2
      has_conflict=1
      continue
    fi

    if [[ -e "${link_path}" ]]; then
      echo "conflict: ${link_path} already exists and is not a symlink" >&2
      has_conflict=1
    fi
  done
done

if [[ ${has_conflict} -ne 0 ]]; then
  echo "aborting without changes because conflicts were found" >&2
  exit 1
fi

for target_dir in "${targets[@]}"; do
  mkdir -p "${target_dir}"

  for skill_path in "${skill_paths[@]}"; do
    skill_name="$(basename "${skill_path}")"
    link_path="${target_dir}/${skill_name}"

    if [[ -L "${link_path}" ]]; then
      echo "exists: ${link_path} -> ${skill_path}"
      continue
    fi

    ln -s "${skill_path}" "${link_path}"
    echo "linked: ${link_path} -> ${skill_path}"
  done
done
