"""Promote release aliases by numeric CalVer order after publishing release assets."""
import json
import os
import re
import subprocess

IMAGE = 'ghcr.io/pipapoen/pipapoen-node-descriptor'
REPOSITORY = 'Pipapoen/pipapoen-node-descriptor'
STABLE = re.compile(r'([0-9]{2})\.([1-9][0-9]*)\.(0|[1-9][0-9]*)')
CANDIDATE = re.compile(STABLE.pattern + r'-rc\.[1-9][0-9]*')


def stable_version(value):
    match = STABLE.fullmatch(value)
    return tuple(map(int, match.groups())) if match else None


def aliases_for(version, releases):
    if CANDIDATE.fullmatch(version):
        return []
    current = stable_version(version)
    if current is None:
        raise ValueError('Invalid CalVer release version')
    published = set()
    for release in releases:
        if release['draft'] or release['prerelease']:
            continue
        tag = release['tag_name']
        parsed = stable_version(tag[1:]) if tag.startswith('v') else None
        if parsed:
            published.add(parsed)
    if current not in published:
        raise ValueError('The stable release must be published before promoting image aliases')
    year, series, _ = current
    aliases = []
    if current == max(v for v in published if v[:2] == current[:2]):
        aliases.append(f'{year:02d}.{series}')
    if current == max(v for v in published if v[0] == year):
        aliases.append(f'{year:02d}')
    if current == max(published):
        aliases.append('latest')
    return aliases


def promote(version, digest):
    if CANDIDATE.fullmatch(version):
        print('Candidate image retains its full version tag only.')
        return []
    if stable_version(version) is None or not re.fullmatch(r'sha256:[0-9a-f]{64}', digest):
        raise ValueError('Expected a stable CalVer version and an image digest')
    # Pagination includes old release lines and compares versions independently of publication time.
    result = subprocess.run(
        ['gh', 'api', '--paginate', '--slurp', f'repos/{REPOSITORY}/releases?per_page=100'],
        check=True, capture_output=True, text=True,
    )
    releases = [release for page in json.loads(result.stdout) for release in page]
    aliases = aliases_for(version, releases)
    if aliases:
        command = ['docker', 'buildx', 'imagetools', 'create', '--prefer-index=false']
        for alias in aliases:
            command.extend(['--tag', f'{IMAGE}:{alias}'])
        command.append(f'{IMAGE}@{digest}')
        subprocess.run(command, check=True)
        if 'latest' in aliases:
            subprocess.run(['gh', 'release', 'edit', 'v' + version, '--repo', REPOSITORY, '--latest'], check=True)
    print('Updated image aliases: ' + (', '.join(aliases) or 'none (newer releases already exist)'))
    return aliases


if __name__ == '__main__':
    promote(os.environ['RELEASE_VERSION'], os.environ['IMAGE_DIGEST'])
